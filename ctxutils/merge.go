package ctxutils

import (
	"context"
	"fmt"
)

func Merge(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {

	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)
	ctx, cancelCauseFn := context.WithCancelCause(ctx)

	go func() {
		for {
			select {
			case <-ctx1.Done():
				cancelCauseFn(fmt.Errorf("%s: %w", name(ctx1, "context-1"), ctx1.Err()))

			case <-ctx2.Done():
				cancelCauseFn(fmt.Errorf("%s: %w", name(ctx2, "context-2"), ctx2.Err()))
			}
		}
	}()
	return ctx, cancelFn
}

func name(ctx context.Context, dflt string) string {
	n, ok := ctx.Value("name").(string)
	if !ok {
		n = dflt
	}
	return n
}
