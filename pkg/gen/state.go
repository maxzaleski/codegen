package gen

import (
	"context"

	"github.com/codegen/internal"
)

// state represents the shared generator state.
type state struct {
	paths   *paths
	metrics *metrics
}

type paths struct {
	// CodegenPath is the path to the `.codegen` directory.
	CodegenPath string
	// PkgOutPath is the path to the output directory hosting the generated packages.
	PkgOutPath string
}

const stateInContextKey internal.ContextKey = "gen_state"

// getStateFromContext returns the generator state from the given context.
func getStateFromContext(ctx context.Context) *state {
	state, ok := ctx.Value(stateInContextKey).(*state)
	if !ok {
		panic("missing generator state in run context")
	}
	return state
}
