package worker

import "sync"

type CancellationCommand struct {
	JobID  uint64
	Reason string
}

// A CancellationBroadcaster allows you to subscribe to and unsubscribe from
// cancellation messages for a given job ID.
type CancellationBroadcaster struct {
	registryMutex sync.Mutex
	registry      map[uint64][](chan CancellationCommand)
}

// NewCancellationBroadcaster sets up a new cancellation broadcaster with an
// empty registry.
func NewCancellationBroadcaster() *CancellationBroadcaster {
	return &CancellationBroadcaster{
		registry: make(map[uint64][](chan CancellationCommand)),
	}
}

// Broadcast broacasts a cancellation message to all currently subscribed
// cancellers.
func (cb *CancellationBroadcaster) Broadcast(command CancellationCommand) {
	cb.registryMutex.Lock()
	defer cb.registryMutex.Unlock()

	chans := cb.registry[command.JobID]
	delete(cb.registry, command.JobID)

	for _, ch := range chans {
		ch <- command
		close(ch)
	}
}

// Subscribe will set up a subscription for cancellation messages for the
// given job ID. When a cancellation message comes in, the returned channel
// will be closed.
func (cb *CancellationBroadcaster) Subscribe(id uint64) <-chan CancellationCommand {
	cb.registryMutex.Lock()
	defer cb.registryMutex.Unlock()

	if _, ok := cb.registry[id]; !ok {
		cb.registry[id] = make([](chan CancellationCommand), 0, 1)
	}

	ch := make(chan CancellationCommand, 1)
	cb.registry[id] = append(cb.registry[id], ch)

	return ch
}

// Unsubscribe removes an existing subscription for the channel.
func (cb *CancellationBroadcaster) Unsubscribe(id uint64, ch <-chan CancellationCommand) {
	cb.registryMutex.Lock()
	defer cb.registryMutex.Unlock()

	// If there's no registered channels for the given ID, just return
	if _, ok := cb.registry[id]; !ok {
		return
	}

	// If there's only one element, remove the key
	if len(cb.registry[id]) <= 1 {
		delete(cb.registry, id)
		return
	}

	var chanIndex int = -1
	for i, registeredChan := range cb.registry[id] {
		if registeredChan == ch {
			chanIndex = i
			break
		}
	}
	if chanIndex == -1 {
		// Channel is already removed
		return
	}

	// Remove element at index by putting the last element in that place, and
	// then shrinking the slice to remove the last element.
	cb.registry[id][chanIndex] = cb.registry[id][len(cb.registry[id])-1]
	cb.registry[id] = cb.registry[id][:len(cb.registry[id])-1]
}
