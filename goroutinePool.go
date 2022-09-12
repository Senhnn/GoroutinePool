package GoroutinePool

import (
	"context"
	"fmt"
	"sync"
)

var defaultPool IPool

var poolMap sync.Map

func init() {
	defaultPool = NewPool("GoroutinePool.DefaultPool", 1000, NewConfig())
}

func Go(f func()) {
	CtxGo(context.Background(), f)
}

func CtxGo(ctx context.Context, f func()) {
	defaultPool.CtxGo(ctx, f)
}

func SetCap(cap int32) {
	defaultPool.SetCap(cap)
}

func SetPanicHandler(f func(context.Context, interface{})) {
	defaultPool.SetPanicHandler(f)
}

func WorkerCount() int32 {
	return defaultPool.WorkerCount()
}

func RegisterPool(p IPool) error {
	_, loaded := poolMap.LoadOrStore(p.Name(), p)
	if loaded {
		return fmt.Errorf("pool name:%s is already registered", p.Name())
	}
	return nil
}

func GetPool(name string) IPool {
	p, ok := poolMap.Load(name)
	if !ok {
		return nil
	}
	return p.(IPool)
}
