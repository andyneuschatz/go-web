package web

import "sync"

// NewCtxPool returns a new CtxPool.
func NewCtxPool(bufferSize int) *CtxPool {
	return &CtxPool{
		Pool: sync.Pool{New: func() interface{} {
			ctx := &Ctx{}
			return ctx
		}},
	}
}

// CtxPool is a sync.Pool of Ctx.
type CtxPool struct {
	sync.Pool
}

// Get returns a pooled bytes.Buffer instance.
func (cp *CtxPool) Get() *Ctx {
	return cp.Pool.Get().(*Ctx)
}

// Put returns the pooled instance.
func (cp *CtxPool) Put(ctx *Ctx) {
	ctx.Reset()
	cp.Pool.Put(ctx)
}
