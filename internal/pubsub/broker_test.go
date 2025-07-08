package pubsub

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试数据结构
type TestMessage struct {
	ID      string
	Content string
}

func TestNewBroker(t *testing.T) {
	t.Parallel()

	broker := NewBroker[TestMessage]()
	assert.NotNil(t, broker)
	
	// 确保broker实现了Suscriber接口
	assert.Implements(t, (*Suscriber[TestMessage])(nil), broker)
}

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker[TestMessage]()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 订阅事件
	eventCh := broker.Subscribe(ctx)
	
	// 发布一个消息
	testMessage := TestMessage{ID: "test-1", Content: "Hello"}
	broker.Publish(CreatedEvent, testMessage)
	
	// 等待接收事件
	select {
	case event := <-eventCh:
		assert.Equal(t, CreatedEvent, event.Type)
		assert.Equal(t, testMessage, event.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event not received within timeout")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker[TestMessage]()
	
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	
	// 两个订阅者
	eventCh1 := broker.Subscribe(ctx1)
	eventCh2 := broker.Subscribe(ctx2)
	
	// 发布消息
	testMessage := TestMessage{ID: "test-multi", Content: "Broadcast"}
	broker.Publish(CreatedEvent, testMessage)
	
	// 验证两个订阅者都收到消息
	var wg sync.WaitGroup
	wg.Add(2)
	
	go func() {
		defer wg.Done()
		select {
		case event := <-eventCh1:
			assert.Equal(t, CreatedEvent, event.Type)
			assert.Equal(t, testMessage, event.Payload)
		case <-time.After(100 * time.Millisecond):
			t.Error("Subscriber 1 did not receive event")
		}
	}()
	
	go func() {
		defer wg.Done()
		select {
		case event := <-eventCh2:
			assert.Equal(t, CreatedEvent, event.Type)
			assert.Equal(t, testMessage, event.Payload)
		case <-time.After(100 * time.Millisecond):
			t.Error("Subscriber 2 did not receive event")
		}
	}()
	
	wg.Wait()
}

func TestBroker_DifferentEvents(t *testing.T) {
	broker := NewBroker[TestMessage]()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	eventCh := broker.Subscribe(ctx)
	
	// 发布不同类型的事件
	createMsg := TestMessage{ID: "create-1", Content: "Created"}
	updateMsg := TestMessage{ID: "update-1", Content: "Updated"}
	deleteMsg := TestMessage{ID: "delete-1", Content: "Deleted"}
	
	broker.Publish(CreatedEvent, createMsg)
	broker.Publish(UpdatedEvent, updateMsg)
	broker.Publish(DeletedEvent, deleteMsg)
	
	// 验证接收到所有事件
	var events []Event[TestMessage]
	for i := 0; i < 3; i++ {
		select {
		case event := <-eventCh:
			events = append(events, event)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Event %d not received within timeout", i+1)
		}
	}
	
	assert.Len(t, events, 3)
	
	// 验证事件类型和内容
	eventTypes := make([]EventType, 3)
	for i, event := range events {
		eventTypes[i] = event.Type
	}
	
	assert.Contains(t, eventTypes, CreatedEvent)
	assert.Contains(t, eventTypes, UpdatedEvent)
	assert.Contains(t, eventTypes, DeletedEvent)
}

func TestBroker_ConcurrentPublish(t *testing.T) {
	broker := NewBroker[TestMessage]()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	eventCh := broker.Subscribe(ctx)
	
	// 并发发布多个消息
	const numMessages = 10
	var wg sync.WaitGroup
	
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			message := TestMessage{
				ID:      fmt.Sprintf("concurrent-%d", id),
				Content: fmt.Sprintf("Message %d", id),
			}
			broker.Publish(CreatedEvent, message)
		}(i)
	}
	
	wg.Wait()
	
	// 接收所有消息
	var receivedCount int
	timeout := time.After(1 * time.Second)
	
	for receivedCount < numMessages {
		select {
		case <-eventCh:
			receivedCount++
		case <-timeout:
			t.Fatalf("Only received %d out of %d messages", receivedCount, numMessages)
		}
	}
	
	assert.Equal(t, numMessages, receivedCount)
}

func TestBroker_NoSubscribers(t *testing.T) {
	t.Parallel()

	broker := NewBroker[TestMessage]()
	
	// 发布消息但没有订阅者 - 不应该panic
	testMessage := TestMessage{ID: "no-sub", Content: "No subscribers"}
	
	assert.NotPanics(t, func() {
		broker.Publish(CreatedEvent, testMessage)
	})
}

func TestBroker_GetSubscriberCount(t *testing.T) {
	broker := NewBroker[TestMessage]()
	
	// 初始状态下没有订阅者
	assert.Equal(t, 0, broker.GetSubscriberCount())
	
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	
	// 添加订阅者
	broker.Subscribe(ctx1)
	assert.Equal(t, 1, broker.GetSubscriberCount())
	
	broker.Subscribe(ctx2)
	assert.Equal(t, 2, broker.GetSubscriberCount())
	
	// 取消订阅
	cancel1()
	time.Sleep(10 * time.Millisecond) // 等待清理
	assert.Equal(t, 1, broker.GetSubscriberCount())
	
	cancel2()
	time.Sleep(10 * time.Millisecond) // 等待清理
	assert.Equal(t, 0, broker.GetSubscriberCount())
}

func TestBroker_Shutdown(t *testing.T) {
	broker := NewBroker[TestMessage]()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	eventCh := broker.Subscribe(ctx)
	
	// 关闭broker
	broker.Shutdown()
	
	// 发布消息应该不会被接收
	testMessage := TestMessage{ID: "after-shutdown", Content: "Should not receive"}
	broker.Publish(CreatedEvent, testMessage)
	
	// 验证channel已关闭
	select {
	case _, ok := <-eventCh:
		assert.False(t, ok, "Event channel should be closed after shutdown")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event channel should be closed immediately after shutdown")
	}
}

func TestEventConstants(t *testing.T) {
	t.Parallel()

	// 验证事件常量
	assert.Equal(t, EventType("created"), CreatedEvent)
	assert.Equal(t, EventType("updated"), UpdatedEvent)
	assert.Equal(t, EventType("deleted"), DeletedEvent)
}