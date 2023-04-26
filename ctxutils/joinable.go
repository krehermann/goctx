package ctxutils

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type JoinableCtx interface {
	Id() string
	context.Context
	//	JoinWith(jctx ...JoinableCtx) JoinableCtx
	Get(string) JoinableCtx
	Keys() []string
	Parents() []JoinableCtx
}

type joinableCtx struct {
	id string
	context.Context

	mu             sync.RWMutex
	wrapped        map[string]JoinableCtx
	orderedParents []JoinableCtx
}

func NewJoinableCtx(name string, ctx context.Context) *joinableCtx {
	return &joinableCtx{
		id:             name,
		Context:        ctx,
		wrapped:        make(map[string]JoinableCtx),
		orderedParents: make([]JoinableCtx, 0),
	}
}

func (j *joinableCtx) Value(key any) any {
	j.mu.RLock()
	var val any
	defer j.mu.RUnlock()
	val = j.Context.Value(key)
	if val != nil {
		return val
	}
	// should be breadth first...
	queue := j.Parents()

	for _, jctx := range queue {
		val = jctx.Value(key)
		if val != nil {
			return val
		}
		queue = append(queue, jctx.Parents()...)
	}
	return nil
}

func (j *joinableCtx) Parents() []JoinableCtx {
	parents := make([]JoinableCtx, 0)
	j.mu.RLock()
	defer j.mu.RUnlock()
	for _, p := range j.orderedParents {
		parents = append(parents, p)
	}
	return parents
}

func (j *joinableCtx) Id() string {
	return j.id
}

func (j *joinableCtx) Get(name string) JoinableCtx {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.wrapped[name]
}

func (j *joinableCtx) Keys() []string {
	keys := make([]string, 0)
	j.mu.RLock()
	defer j.mu.RUnlock()
	for k := range j.wrapped {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func Join(name string, jctxs ...JoinableCtx) (JoinableCtx, context.CancelFunc, error) {
	if len(jctxs) < 2 {
		return nil, nil, fmt.Errorf("Join requires at least two input contexts")
	}

	ctxRoot := context.Background()
	ctxInternal, cancelCauseFn := context.WithCancelCause(ctxRoot)
	//ctxExternal, cancelFn := context.WithCancel(ctxInternal)

	out := NewJoinableCtx(name, ctxInternal)

	out.mu.Lock()
	for _, jctx := range jctxs {
		out.wrapped[jctx.Id()] = jctx
		out.orderedParents = append(out.orderedParents, jctx)
	}
	out.mu.Unlock()

	for _, jctx := range jctxs {
		go func(jctx JoinableCtx) {
			select {
			case <-jctx.Done():
				err := fmt.Errorf("%s: %w", jctx.Id(), jctx.Err())
				cancelCauseFn(err)

			case <-ctxInternal.Done():
			}
		}(jctx)
	}

	return out, func() { cancelCauseFn(nil) }, nil
}
