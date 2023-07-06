package gen

import (
	"context"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/pkg/errors"
	"os"
	"strings"
	"sync"
)

func worker(ctx context.Context, id int, wg *sync.WaitGroup, q IQueue, errChan chan<- error) {
	defer wg.Done()

	pkgs, m, l :=
		ctx.Value(contextKeyPackages).([]*core.Package),
		ctx.Value(contextKeyMetrics).(*metrics.Metrics),
		newLogger(
			ctx.Value(contextKeyLogger).(slog.ILogger),
			fmt.Sprintf("worker_%d", id),
			slog.None,
		)

	l.Log("start", "msg", "starting worker")
	defer l.Log("exit", "msg", "worker exiting")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			code := func() (code int) {
				j := q.Dequeue(id)
				if j == nil {
					return -1 // queue is empty; all jobs processed.
				}

				setFileAbsolutePath(j)
				l.Ack("ack<-", j)

				o, sk :=
					&metrics.FileOutcome{AbsolutePath: j.FileAbsolutePath},
					j.Metadata.ScopeKey
				if s := strings.Split(sk, "/"); len(s) == 2 {
					sk = s[0]
				}
				defer func() {
					tag := core.UniquePkgAlias
					if p := j.Package; p != nil {
						tag = p.Name
					}
					m.CaptureScope(sk, tag, *o)
				}()

				if err := generateFile(pkgs, l, j); err != nil {
					if errors.Is(errFileAlreadyPresent, err) {
						return
					}

					errChan <- err
					return -1
				}
				defer logFileOutcome(l, fileOutcomeSuccess, j)

				o.Created = true
				return
			}()
			if code != 0 {
				return
			}
		}
	}
}

var errFileAlreadyPresent = errors.New("file already exists")

func generateFile(pkgs []*core.Package, l ILogger, j *genJob) error {
	if _, err := os.Stat(j.FileAbsolutePath); err != nil {
		if !os.IsNotExist(err) {
			return errors.WithMessagef(err, "failed presence check at '%s'", j.FileAbsolutePath)
		}
	} else {
		defer logFileOutcome(l, fileOutcomeIgnored, j)
		return errFileAlreadyPresent
	}

	return (templateFactory{
		j,
		pkgs,
		nil}).ExecuteTemplate()
}

func setFileAbsolutePath(j *genJob) {
	oc, pkg, fn := j.Metadata, j.Package, j.FileName

	// Apply modifiers to file name.
	for _, mod := range fn.Mods {
		token, pm, sm := mod.Token, core.CaseModifierNone, core.CaseModifierNone
		if token == "pkg" && pkg != nil {
			token = pkg.Name
		}
		for _, m := range mod.Modifiers {
			if core.PrimaryCaseModifier(m).IsValid() {
				pm = m
			} else if core.SecondaryCaseModifier(m).IsValid() {
				sm = m
			}
		}
		key := fmt.Sprintf("{%d}", mod.Key)
		fn.Value = strings.Replace(fn.Value, key, applyCaseModifiers(token, pm, sm), 1)
	}

	j.FileAbsolutePath = oc.AbsolutePath + "/"

	// Inline: files are generated within the same directory space (e.g. models > User.Java, Car.Java).
	// Unique: job is only to be performed once for the specified output.
	if oc.Inline || j.Unique {
		j.FileAbsolutePath += fn.Value
	} else {
		j.FileAbsolutePath += pkg.Name + "/" + fn.Value
	}
}

func applyCaseModifiers(token string, pm core.CaseModifier, sm core.CaseModifier) (result string) {
	if pm == core.CaseModifierLower && sm == core.CaseModifierCamel {
		result = strcase.ToLowerCamel(token)
	} else if pm == core.CaseModifierUpper && sm == core.CaseModifierSnake {
		result = strcase.ToScreamingSnake(token)
	} else if pm == core.CaseModifierUpper && sm == core.CaseModifierKebab {
		result = strcase.ToScreamingKebab(token)
	} else if pm == core.CaseModifierLower && sm == core.CaseModifierSnake ||
		pm == core.CaseModifierNone && sm == core.CaseModifierSnake {
		result = strcase.ToSnake(token)
	} else if pm == core.CaseModifierLower && sm == core.CaseModifierKebab ||
		pm == core.CaseModifierNone && sm == core.CaseModifierKebab {
		result = strcase.ToKebab(token)
	} else if pm == core.CaseModifierTitle ||
		pm == core.CaseModifierNone && sm == core.CaseModifierCamel {
		result = strcase.ToCamel(token)
	} else {
		result = token
	}
	return
}
