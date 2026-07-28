package goak

import (
	"context"
	"errors"
)

var (
	// ErrAppStopped is returned when work is dispatched after application
	// shutdown has started.
	ErrAppStopped = errors.New("goak: application is stopped")

	// ErrNilDispatch is returned when Dispatch receives a nil function.
	ErrNilDispatch = errors.New("goak: dispatched function is nil")
)

type dispatchEntry struct {
	key    string
	latest bool
	run    func()
}

// Context returns the application lifetime context. It is canceled when Run
// exits or Destroy is called.
func (a *App) Context() context.Context {
	a.dispatchMu.Lock()
	defer a.dispatchMu.Unlock()
	a.ensureContextLocked()
	return a.ctx
}

// Dispatch schedules a function to run on the SDL/UI thread before the next
// layout and draw pass. It is safe to call from any goroutine.
func (a *App) Dispatch(run func()) error {
	return a.enqueue(dispatchEntry{run: run})
}

// DispatchLatest schedules keyed work on the SDL/UI thread. If work with the
// same key is already pending, it is replaced. This prevents high-frequency
// streams from building an obsolete update backlog.
func (a *App) DispatchLatest(key string, run func()) error {
	if key == "" {
		return a.Dispatch(run)
	}
	return a.enqueue(dispatchEntry{key: key, latest: true, run: run})
}

func (a *App) enqueue(entry dispatchEntry) error {
	if entry.run == nil {
		return ErrNilDispatch
	}

	a.dispatchMu.Lock()
	defer a.dispatchMu.Unlock()
	a.ensureContextLocked()
	if a.stopped {
		return ErrAppStopped
	}

	if entry.latest {
		if index, ok := a.latestDispatch[entry.key]; ok {
			a.dispatchQueue[index].run = entry.run
			return nil
		}
		a.latestDispatch[entry.key] = len(a.dispatchQueue)
	}
	a.dispatchQueue = append(a.dispatchQueue, entry)
	return nil
}

func (a *App) drainDispatch() {
	a.dispatchMu.Lock()
	queue := a.dispatchQueue
	a.dispatchQueue = nil
	clear(a.latestDispatch)
	a.dispatchMu.Unlock()

	for _, entry := range queue {
		entry.run()
	}
}

func (a *App) stopDispatch() {
	a.dispatchMu.Lock()
	a.ensureContextLocked()
	if a.stopped {
		a.dispatchMu.Unlock()
		return
	}
	a.stopped = true
	a.dispatchQueue = nil
	clear(a.latestDispatch)
	cancel := a.cancel
	a.dispatchMu.Unlock()
	cancel()
}

func (a *App) ensureContextLocked() {
	if a.ctx != nil {
		return
	}
	a.ctx, a.cancel = context.WithCancel(context.Background())
	if a.latestDispatch == nil {
		a.latestDispatch = make(map[string]int)
	}
}
