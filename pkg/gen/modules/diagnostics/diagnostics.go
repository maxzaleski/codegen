package diagnostics

import (
	"context"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/db"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/mitchellh/hashstructure/v2"
	"golang.org/x/sync/errgroup"
)

type (
	IDiagnostics interface {
		// Prepare prepares the diagnostics module for utilisation.
		Prepare(spec *core.Spec) error
		// Verify checks whether a package has changed based on the override rules.
		Verify(map[string]core.ScopeJobOverride) (bool, error)
	}

	hasUpdated = core.ScopeJobOverride

	diagnostics struct {
		logger     slog.INamedLogger
		repository IRepository

		resultsMap     map[string]*core.ScopeJobOverride
		pkgsMap        map[string]core.Package
		pkgsLastModMap map[string]int64
	}

	snapshot struct {
		LastModified int64
		Hash         int64
	}
)

func New(logger slog.ILogger, db db.IDatabase) IDiagnostics {
	return &diagnostics{
		logger:     slog.NewNamed(logger, "diagnostics", slog.None),
		repository: newRepository(logger, db),
		resultsMap: map[string]*core.ScopeJobOverride{},
		pkgsMap:    map[string]core.Package{},
	}
}

func (d *diagnostics) Prepare(spec *core.Spec) error {
	d.logger.Log("prepare", "msg", "preparing diagnostics module")

	// -> Seed the database.
	if err := d.repository.SeedDB(); err != nil {
		return err
	}

	// -> Prepare the diagnostics module.
	d.pkgsLastModMap = spec.Metadata.PkgsLastModifiedMap
	for _, pkg := range spec.Pkgs {
		d.pkgsMap[pkg.Name] = *pkg
	}
	return nil
}

func (d *diagnostics) Verify(overrOn map[string]core.ScopeJobOverride) (bool, error) {
	_ = func(msg string, lines ...any) {
		linesCopy := make([]any, 2+len(lines))
		linesCopy = append(linesCopy, "msg", msg)
		linesCopy = append(linesCopy, lines...)
		d.logger.Log("verify", linesCopy...)
	}
	hasChanged := func(pkg string, hasChanged, overr core.ScopeJobOverride) bool {
		if overr.Model && hasChanged.Model {
			return true
		}
		if overr.Interface && hasChanged.Interface {
			return true
		}
		return false
	}

	for pkg, overr := range overrOn {
		// -> Check if the package has already been checked through snapshotting.
		if hc, ok := d.resultsMap[pkg]; ok {
			if ok = hasChanged(pkg, *hc, overr); ok {
				return true, nil
			}
		} else {
			// -> Perform snapshotting.
			if err := d.performSnapshot(pkg, overr); err != nil {
				return false, err
			}
			if ok = hasChanged(pkg, *d.resultsMap[pkg], overr); ok {
				return true, nil
			}
		}
	}
	return false, nil // No changes detected.
}

func (d *diagnostics) performSnapshot(pkgS string, overr core.ScopeJobOverride) error {
	d.resultsMap[pkgS] = &core.ScopeJobOverride{}

	performSnapshot := func(ctx context.Context, pi int) error {
		// This function is responsible for comparing the old and current snapshots for a given package and property.
		// If the snapshots are different, the property has changed and the function will return true.
		hasChanged := func(pi int) (bool, error) {
			// [1] Get old snapshot.
			s, err := d.repository.FindOne(ctx, pkgS, pi)
			if err != nil {
				return false, err
			}

			pkg, nS :=
				d.pkgsMap[pkgS],
				&snapshot{LastModified: d.pkgsLastModMap[pkgS]}

			// [2] Compare snapshots.
			hasChanged := false
			if s.LastModified != nS.LastModified {
				// -> Perform local snapshot.
				var hash uint64
				switch pi {
				case 0:
					hash, err = d.hash(pkg.Models)
				case 1:
					hash, err = d.hash(pkg.Interface)
				}
				if err != nil {
					return false, err
				}
				nS.Hash = int64(hash)
				// -> Compare hashes.
				hasChanged = s.Hash != nS.Hash
			}
			if pkgS != "\\*" {
				if err = d.repository.InsertOne(ctx, pkgS, pi, *nS); err != nil {
					return false, err
				}
			}

			return hasChanged, nil
		}
		if ok, err := hasChanged(pi); err != nil {
			return err
		} else {
			if ok {
				d.resultsMap[pkgS].Set(pi, true)
			}
			return nil
		}
	}

	// -> Compare for each available property.
	errg, ctx := errgroup.WithContext(context.Background())
	for i, ok := range overr.AsSlice() {
		if ok {
			errg.Go(func() error { return performSnapshot(ctx, i) })
		}
	}

	return errg.Wait()
}

func (d *diagnostics) hash(v interface{}) (uint64, error) {
	opts := &hashstructure.HashOptions{SlicesAsSets: true}
	return hashstructure.Hash(v, hashstructure.FormatV2, opts)
}
