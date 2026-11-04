package server

import (
	"testing"
	"time"

	"github.com/neouks/series/db"
)

// drainUntilClosed consumes whatever is buffered and returns how many events were
// still readable before the channel reported closed.
func drainUntilClosed(t *testing.T, ch <-chan db.Activity) int {
	t.Helper()
	deadline := time.After(3 * time.Second)
	got := 0
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return got
			}
			got++
		case <-deadline:
			t.Fatalf("channel still open after %d buffered events", got)
		}
	}
}

// A subscriber that stops reading must be disconnected (channel closed) instead of
// silently losing events: the SSE handler returns and the browser replays the gap
// from its Last-Event-ID.
func TestBroadcastOverflowClosesSlowSubscriber(t *testing.T) {
	b := NewBroadcaster()
	slow, unsubSlow := b.Subscribe("t1")
	fast, unsubFast := b.Subscribe("t1")
	defer unsubFast()

	fastGot := make(chan int, 1)
	go func() {
		n := 0
		for range fast {
			n++
		}
		fastGot <- n
	}()

	const total = 300 // > 256 buffer, so the non-reading subscriber must overflow
	for i := 0; i < total; i++ {
		b.Publish("t1", db.Activity{ID: int64(i + 1)})
	}

	if n := drainUntilClosed(t, slow); n == 0 {
		t.Fatal("slow subscriber was closed before receiving anything")
	}

	// Publishing again after the disconnect must not panic on the closed channel.
	b.Publish("t1", db.Activity{ID: 9999})

	unsubFast()
	select {
	case n := <-fastGot:
		if n == 0 {
			t.Fatal("active subscriber received nothing")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("unsubscribe did not close the active subscriber")
	}

	// The handler's deferred unsubscribe runs after Publish already closed it.
	unsubSlow()

	b.mu.Lock()
	left := len(b.subs)
	b.mu.Unlock()
	if left != 0 {
		t.Fatalf("subscriptions left behind: %d", left)
	}
}

func TestBroadcastDeliversInOrder(t *testing.T) {
	b := NewBroadcaster()
	ch, unsub := b.Subscribe("t2")
	defer unsub()

	const total = 50
	go func() {
		for i := 1; i <= total; i++ {
			b.Publish("t2", db.Activity{ID: int64(i)})
		}
	}()
	for i := 1; i <= total; i++ {
		select {
		case a := <-ch:
			if a.ID != int64(i) {
				t.Fatalf("event %d: got id %d", i, a.ID)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}
}

func TestBroadcastUnsubscribeAfterOverflowIsSafe(t *testing.T) {
	b := NewBroadcaster()
	_, unsub := b.Subscribe("t3")
	for i := 0; i < 300; i++ {
		b.Publish("t3", db.Activity{ID: int64(i + 1)})
	}
	unsub()
	unsub() // idempotent: must not panic on an already-closed channel

	b.mu.Lock()
	left := len(b.subs)
	b.mu.Unlock()
	if left != 0 {
		t.Fatalf("subscriptions left behind: %d", left)
	}
}
