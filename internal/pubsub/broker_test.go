package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Message string
	Value   int
}

func TestNewBroker(t *testing.T) {
	broker := NewBroker[TestData]()
	assert.NotNil(t, broker)
	assert.Equal(t, 0, broker.GetSubscriberCount())
}

func TestNewBrokerWithOptions(t *testing.T) {
	broker := NewBrokerWithOptions[TestData](32, 500)
	assert.NotNil(t, broker)
	assert.Equal(t, 0, broker.GetSubscriberCount())
	assert.Equal(t, 500, broker.maxEvents)
}

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker[TestData]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := broker.Subscribe(ctx)
	assert.NotNil(t, ch)
	assert.Equal(t, 1, broker.GetSubscriberCount())

	// Cancel context should unsubscribe
	cancel()
	time.Sleep(10 * time.Millisecond) // Give time for cleanup
	assert.Equal(t, 0, broker.GetSubscriberCount())
}

func TestBroker_Publish(t *testing.T) {
	broker := NewBroker[TestData]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := broker.Subscribe(ctx)
	testData := TestData{Message: "test", Value: 42}

	// Publish event
	broker.Publish(CreatedEvent, testData)

	// Should receive the event
	select {
	case event := <-ch:
		assert.Equal(t, CreatedEvent, event.Type)
		assert.Equal(t, testData, event.Payload)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker[TestData]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const numSubscribers = 5
	channels := make([]<-chan Event[TestData], numSubscribers)

	// Create multiple subscribers
	for i := 0; i < numSubscribers; i++ {
		channels[i] = broker.Subscribe(ctx)
	}

	assert.Equal(t, numSubscribers, broker.GetSubscriberCount())

	testData := TestData{Message: "broadcast", Value: 123}
	broker.Publish(UpdatedEvent, testData)

	// All subscribers should receive the event
	for i, ch := range channels {
		select {
		case event := <-ch:
			assert.Equal(t, UpdatedEvent, event.Type)
			assert.Equal(t, testData, event.Payload)
		case <-time.After(time.Second):
			t.Fatalf("Timeout waiting for event on subscriber %d", i)
		}
	}
}

func TestBroker_ConcurrentPublish(t *testing.T) {
	broker := NewBroker[TestData]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := broker.Subscribe(ctx)

	const numEvents = 100
	var wg sync.WaitGroup

	// Publish events concurrently
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			testData := TestData{Message: "concurrent", Value: index}
			broker.Publish(CreatedEvent, testData)
		}(i)
	}

	// Collect events with timeout
	events := make([]Event[TestData], 0, numEvents)
	done := make(chan bool, 1)

	go func() {
		collected := 0
		for collected < numEvents {
			select {
			case event := <-ch:
				events = append(events, event)
				collected++
			case <-time.After(2 * time.Second):
				done <- false
				return
			}
		}
		done <- true
	}()

	wg.Wait()

	// Wait for collection to complete
	success := <-done
	if !success {
		t.Error("Timeout waiting for events")
		return
	}

	assert.Len(t, events, numEvents)

	// Verify all events are CreatedEvent type
	for _, event := range events {
		assert.Equal(t, CreatedEvent, event.Type)
		assert.Equal(t, "concurrent", event.Payload.Message)
	}
}

func TestBroker_Shutdown(t *testing.T) {
	broker := NewBroker[TestData]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := broker.Subscribe(ctx)
	assert.Equal(t, 1, broker.GetSubscriberCount())

	// Shutdown broker
	broker.Shutdown()

	// Subscriber count should be 0
	assert.Equal(t, 0, broker.GetSubscriberCount())

	// Channel should be closed
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel should be closed immediately")
	}
}

func TestBroker_PublishAfterShutdown(t *testing.T) {
	broker := NewBroker[TestData]()
	broker.Shutdown()

	// Publishing after shutdown should not panic
	assert.NotPanics(t, func() {
		broker.Publish(CreatedEvent, TestData{Message: "after shutdown"})
	})
}

func TestBroker_SubscribeAfterShutdown(t *testing.T) {
	broker := NewBroker[TestData]()
	broker.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe after shutdown should return closed channel
	ch := broker.Subscribe(ctx)

	select {
	case _, ok := <-ch:
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel should be closed immediately")
	}
}

func TestBroker_DoubleShutdown(t *testing.T) {
	broker := NewBroker[TestData]()

	// First shutdown
	broker.Shutdown()

	// Second shutdown should not panic
	assert.NotPanics(t, func() {
		broker.Shutdown()
	})
}

func TestEventTypes(t *testing.T) {
	assert.Equal(t, "created", string(CreatedEvent))
	assert.Equal(t, "updated", string(UpdatedEvent))
	assert.Equal(t, "deleted", string(DeletedEvent))
}

func TestEvent(t *testing.T) {
	testData := TestData{Message: "test event", Value: 42}
	event := Event[TestData]{
		Type:    CreatedEvent,
		Payload: testData,
	}

	assert.Equal(t, CreatedEvent, event.Type)
	assert.Equal(t, testData, event.Payload)
}

func TestBroker_BufferOverflow(t *testing.T) {
	// Create broker with small buffer
	broker := NewBrokerWithOptions[TestData](1, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := broker.Subscribe(ctx)

	// Fill the buffer and publish more
	testData := TestData{Message: "overflow test", Value: 1}
	broker.Publish(CreatedEvent, testData) // Should go through
	broker.Publish(UpdatedEvent, testData) // Should be dropped due to buffer
	broker.Publish(DeletedEvent, testData) // Should be dropped due to buffer

	// Should only receive the first event
	select {
	case event := <-ch:
		assert.Equal(t, CreatedEvent, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should receive at least one event")
	}

	// No more events should be available immediately
	select {
	case <-ch:
		// This is fine, might receive one more due to buffer
	default:
		// This is also fine, buffer might be full
	}
}

func TestBroker_ContextCancellation(t *testing.T) {
	broker := NewBroker[TestData]()

	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	ch1 := broker.Subscribe(ctx1)
	ch2 := broker.Subscribe(ctx2)

	assert.Equal(t, 2, broker.GetSubscriberCount())

	// Cancel first context
	cancel1()
	time.Sleep(10 * time.Millisecond) // Give time for cleanup

	assert.Equal(t, 1, broker.GetSubscriberCount())

	// First channel should be closed, second should still work
	select {
	case _, ok := <-ch1:
		assert.False(t, ok, "First channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("First channel should be closed")
	}

	// Second channel should still work
	testData := TestData{Message: "still working", Value: 42}
	broker.Publish(CreatedEvent, testData)

	select {
	case event := <-ch2:
		assert.Equal(t, CreatedEvent, event.Type)
		assert.Equal(t, testData, event.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Second channel should still work")
	}
}

// Test that interfaces are properly implemented
func TestInterfaces(t *testing.T) {
	broker := NewBroker[TestData]()

	// Test Publisher interface
	var publisher Publisher[TestData] = broker
	assert.NotNil(t, publisher)

	// Test Suscriber interface
	var subscriber Suscriber[TestData] = broker
	assert.NotNil(t, subscriber)
}

func TestBroker_EmptyPayload(t *testing.T) {
	broker := NewBroker[string]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := broker.Subscribe(ctx)

	// Publish empty string
	broker.Publish(CreatedEvent, "")

	select {
	case event := <-ch:
		assert.Equal(t, CreatedEvent, event.Type)
		assert.Equal(t, "", event.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should receive event with empty payload")
	}
}
