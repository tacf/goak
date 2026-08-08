package goak

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestDispatchRunsInQueueOrder(t *testing.T) {
	app := NewApp()
	defer app.Destroy()

	var got []int
	for i := 1; i <= 3; i++ {
		value := i
		if err := app.Dispatch(func() { got = append(got, value) }); err != nil {
			t.Fatal(err)
		}
	}

	app.drainDispatch()
	if want := []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("dispatch order = %v, want %v", got, want)
	}
}

func TestDispatchIsSafeFromConcurrentProducers(t *testing.T) {
	app := NewApp()
	defer app.Destroy()

	const count = 256
	var producers sync.WaitGroup
	for range count {
		producers.Add(1)
		go func() {
			defer producers.Done()
			if err := app.Dispatch(func() {}); err != nil {
				t.Errorf("Dispatch() error = %v", err)
			}
		}()
	}
	producers.Wait()

	app.dispatchMu.Lock()
	queued := len(app.dispatchQueue)
	app.dispatchMu.Unlock()
	if queued != count {
		t.Fatalf("queued actions = %d, want %d", queued, count)
	}
}

func TestDispatchLatestCoalescesByKey(t *testing.T) {
	app := NewApp()
	defer app.Destroy()

	var got []string
	_ = app.Dispatch(func() { got = append(got, "first") })
	_ = app.DispatchLatest("progress", func() { got = append(got, "stale") })
	_ = app.Dispatch(func() { got = append(got, "middle") })
	_ = app.DispatchLatest("progress", func() { got = append(got, "latest") })
	_ = app.DispatchLatest("status", func() { got = append(got, "status") })

	app.drainDispatch()
	want := []string{"first", "latest", "middle", "status"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("coalesced dispatch = %v, want %v", got, want)
	}
}

func TestDispatchDuringDrainRunsOnNextDrain(t *testing.T) {
	app := NewApp()
	defer app.Destroy()

	var count int
	_ = app.Dispatch(func() {
		count++
		_ = app.Dispatch(func() { count++ })
	})

	app.drainDispatch()
	if count != 1 {
		t.Fatalf("count after first drain = %d, want 1", count)
	}
	app.drainDispatch()
	if count != 2 {
		t.Fatalf("count after second drain = %d, want 2", count)
	}
}

func TestDestroyCancelsContextAndRejectsDispatch(t *testing.T) {
	app := NewApp()
	ctx := app.Context()
	app.Destroy()

	select {
	case <-ctx.Done():
	default:
		t.Fatal("application context was not canceled")
	}
	if err := app.Dispatch(func() {}); !errors.Is(err, ErrAppStopped) {
		t.Fatalf("Dispatch() error = %v, want ErrAppStopped", err)
	}
	if err := app.Dispatch(nil); !errors.Is(err, ErrNilDispatch) {
		t.Fatalf("Dispatch(nil) error = %v, want ErrNilDispatch", err)
	}
}
