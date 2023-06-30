package gen

import (
	"context"
	"fmt"
	"github.com/codegen/internal/core"
	"github.com/codegen/internal/core/slog"
	"github.com/codegen/internal/metrics"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"os"
	"strings"
	"sync"
	"time"
)

func worker(ctx context.Context, id int, wg *sync.WaitGroup, q IQueue, errChan chan<- error) {
	defer wg.Done()

	m := ctx.Value("metrics").(*metrics.Metrics)
	l := newLogger(
		ctx.Value("logger").(slog.ILogger),
		ctx.Value("began").(time.Time),
		id,
	)

	l.Start()
	defer l.Exit()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			code := func() (code int) {
				j, ok := q.Dequeue()
				if !ok {
					return -1
				}

				setFileAbsolutePath(j)
				l.Ack(j)

				mrt, sk := &metrics.Measurement{FileAbsolutePath: j.FileAbsolutePath}, j.Metadata.ScopeKey
				if s := strings.Split(sk, "/"); len(s) == 2 {
					sk = s[0]
				}
				defer func() {
					tag := "[unique]"
					if p := j.Package; p != nil {
						tag = p.Name
					}
					m.Measure(sk, tag, mrt)
				}()

				if err := generateFile(l, j); err != nil {
					if errors.Is(errFileAlreadyPresent, err) {
						return
					}

					errChan <- err
					return -1
				}
				defer l.EventFile("created", j)

				mrt.Created = true
				return
			}()
			if code != 0 {
				return
			}
		}
	}
}

var errFileAlreadyPresent = errors.New("file already exists")

func generateFile(wl ilogger, j *genJob) error {
	if _, err := os.Stat(j.FileAbsolutePath); err != nil {
		if !os.IsNotExist(err) {
			return errors.WithMessagef(err, "failed presence check at '%s'", j.FileAbsolutePath)
		}
	} else {
		wl.EventFile("skipped", j)
		return errFileAlreadyPresent
	}

	return (templateFactory{j}).ExecuteTemplate()
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
