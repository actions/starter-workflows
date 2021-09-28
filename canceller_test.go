package worker

import (
	"testing"
)

func TestCancellationBroadcaster(t *testing.T) {
	cb := NewCancellationBroadcaster()

	ch1_1 := cb.Subscribe(1)
	ch1_2 := cb.Subscribe(1)
	ch1_3 := cb.Subscribe(1)
	ch2 := cb.Subscribe(2)

	cb.Unsubscribe(1, ch1_2)

	cb.Broadcast(CancellationCommand{JobID: 1, Reason: "42"})
	cb.Broadcast(CancellationCommand{JobID: 1, Reason: "42"})

	assertReceived(t, "ch1_1", ch1_1, CancellationCommand{JobID: 1, Reason: "42"})
	assertWaiting(t, "ch1_2", ch1_2)
	assertReceived(t, "ch1_3", ch1_3, CancellationCommand{JobID: 1, Reason: "42"})
	assertWaiting(t, "ch2", ch2)
}

func assertReceived(t *testing.T, name string, ch <-chan CancellationCommand, expected CancellationCommand) {
	select {
	case val, ok := (<-ch):
		if ok {
			if expected != val {
				t.Errorf("expected to receive %v, got %v", expected, val)
			}
		} else {
			t.Errorf("expected %s to not be closed, but it was closed", name)
		}
	default:
		t.Errorf("expected %s to receive a value, but it didn't", name)
	}
}

func assertWaiting(t *testing.T, name string, ch <-chan CancellationCommand) {
	select {
	case _, ok := (<-ch):
		if ok {
			t.Errorf("expected %s to not be closed and not have a value, but it received a value", name)
		} else {
			t.Errorf("expected %s to not be closed and not have a value, but it was closed", name)
		}
	default:
	}
}
