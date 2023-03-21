package gen

import (
	"context"
	"testing"
)

func Test_getStateFromContext(t *testing.T) {
	ctx := context.Background()

	t.Run("exists", func(t *testing.T) {
		state := &state{
			paths:   &paths{},
			metrics: &metrics{},
		}
		gotState := getStateFromContext(context.WithValue(ctx, stateInContextKey, state))
		if gotState != state {
			t.Errorf("getStateFromContext() = %v, want %v", gotState, state)
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("getStateFromContext did not panic as expected")
			}
		}()
		getStateFromContext(ctx)
	})
}
