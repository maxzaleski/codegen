package gen

import (
	"context"
	"fmt"
	"github.com/codegen/internal/core"
	"github.com/codegen/internal/metrics"
	"github.com/codegen/pkg/slog"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"os"
	"strings"
	"sync"
)

func worker(ctx context.Context, wg *sync.WaitGroup, q IQueue, errChan chan<- error) {
	defer wg.Done()

	m := ctx.Value("metrics").(*metrics.Metrics)
	l := ctx.Value("logger").(slog.ILogger)

	for j := range q.Stream() {
		mrt := &metrics.Measurement{ScopeKey: j.ScopeKey}

		if err := generateFile(ctx, j); err != nil {
			errChan <- err
			q.Close()
		} else {
			mrt.Created = true
			l.Logf("[%s] created successfully", j.FileName.Value)
		}

		mrt.FileAbsolutePath = j.FileAbsolutePath
		m.Measure(j.Package.Name, mrt)
	}
}

func generateFile(ctx context.Context, j *job) error {
	l := ctx.Value("logger").(slog.ILogger)

	fileAbsolutePath, fn := getFileAbsolutePath(j), j.FileName

	if _, err := os.Stat(fileAbsolutePath); err != nil {
		if !os.IsNotExist(err) {
			return errors.WithMessagef(err, "failed presence check at '%s'", fileAbsolutePath)
		}
		l.Logf("[%s] already present -- skipped", fn.Value)
	}

	tf := templateFactory{j}
	return tf.ExecuteTemplate()
}

func getFileAbsolutePath(j *job) (s string) {
	oc, pkg, fn := j.Metadata, j.Package, j.FileName

	for _, mod := range fn.Mods {
		token, pm, sm := mod.Token, core.CaseModifierNone, core.CaseModifierNone
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

	s = oc.AbsolutePath
	if !oc.Inline {
		s += fmt.Sprintf("/%s/%s", pkg.Name, fn.Value)
	} else {
		s += "/" + fn.Value
	}
	return
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
