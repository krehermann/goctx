package ctxutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJoin(t *testing.T) {

	t.Run("not enough args", func(t *testing.T) {
		jctx, cancelFunc, err := Join("child", NewJoinableCtx("parent", context.Background()))
		require.Error(t, err)
		require.Nil(t, jctx)
		require.Nil(t, cancelFunc)

	})

	t.Run("join with values", func(t *testing.T) {
		ctx := context.Background()
		p1ctx := context.WithValue(ctx, "say", "hello")
		p1ctx, p1cancel := context.WithCancel(p1ctx)

		p2ctx := context.WithValue(context.Background(), "say", "goodbye")
		p2ctx, p2cancel := context.WithCancel(p2ctx)

		jctx, cancelFn, err := Join("derived",
			NewJoinableCtx("p1", p1ctx),
			NewJoinableCtx("p2", p2ctx),
		)
		defer cancelFn()
		require.NoError(t, err)

		p1cancel()
		defer p2cancel()
		<-jctx.Done()

		require.Error(t, jctx.Err())
		cerr := context.Cause(jctx)
		t.Logf("joint err %v", cerr)
		require.Contains(t, cerr.Error(), "p1")

		require.Equal(t, jctx.Id(), "derived")
		require.Equal(t, jctx.Keys(), []string{"p1", "p2"})

		gotp1ctx := jctx.Get("p1")
		val1, ok := gotp1ctx.Value("say").(string)
		require.True(t, ok)
		require.Equal(t, "hello", val1)

		firstVal, ok := jctx.Value("say").(string)
		require.True(t, ok)
		require.Equal(t, "hello", firstVal)
	})

}
