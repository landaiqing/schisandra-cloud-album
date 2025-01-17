package events

import (
	"fmt"
	"reflect"
	"sync"
)

// Event 定义事件类型
type Event struct {
	Name string      // 事件名称
	Data interface{} // 事件数据
}

// EventHandler 定义事件处理器函数类型
type EventHandler func(event Event)

// Dispatcher 接口定义事件分发器
type Dispatcher interface {
	Register(eventName string, handler EventHandler)            // 注册事件处理器
	RegisterOnce(eventName string, handler EventHandler)        // 注册一次性事件处理器
	Dispatch(event Event)                                       // 分发事件
	RemoveHandler(eventName string, handler EventHandler) error // 移除特定处理器
	ClearHandlers(eventName string)                             // 清除某事件的所有处理器
}

// defaultDispatcher 默认事件分发器实现
type defaultDispatcher struct {
	handlers map[string][]EventHandler
	once     map[string]map[*EventHandler]struct{} // 使用指针作为 map 键
	mu       sync.RWMutex
}

// NewDispatcher 创建新的事件分发器
func NewDispatcher() Dispatcher {
	return &defaultDispatcher{
		handlers: make(map[string][]EventHandler),
		once:     make(map[string]map[*EventHandler]struct{}), // 修改为指针
	}
}

// Register 注册事件处理器
func (d *defaultDispatcher) Register(eventName string, handler EventHandler) {
	if eventName == "" || handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[eventName] = append(d.handlers[eventName], handler)
}

// RegisterOnce 注册一次性事件处理器
func (d *defaultDispatcher) RegisterOnce(eventName string, handler EventHandler) {
	if eventName == "" || handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// 如果还未初始化一次性处理器记录表，则初始化
	if _, exists := d.once[eventName]; !exists {
		d.once[eventName] = make(map[*EventHandler]struct{}) // 修改为指针
	}
	d.once[eventName][&handler] = struct{}{}

	// 追加处理器
	d.handlers[eventName] = append(d.handlers[eventName], handler)
}

// Dispatch 分发事件
func (d *defaultDispatcher) Dispatch(event Event) {
	if event.Name == "" {
		return
	}

	d.mu.RLock()
	handlers := d.handlers[event.Name]
	onceHandlers := d.once[event.Name]
	d.mu.RUnlock()

	if len(handlers) == 0 {
		fmt.Printf("No handlers registered for event: %s\n", event.Name)
		return
	}

	var wg sync.WaitGroup

	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			h(event)
		}(handler)
	}

	wg.Wait() // 等待所有处理器执行完毕

	// 移除已执行的一次性处理器
	if len(onceHandlers) > 0 {
		d.mu.Lock()
		defer d.mu.Unlock()

		remainingHandlers := make([]EventHandler, 0, len(handlers))
		for _, handler := range handlers {
			if _, exists := onceHandlers[&handler]; !exists {
				remainingHandlers = append(remainingHandlers, handler)
			} else {
				delete(d.once[event.Name], &handler)
			}
		}
		d.handlers[event.Name] = remainingHandlers
	}
}

// contains 检查事件处理器是否在一次性处理器中
func contains(onceHandlers map[*EventHandler]struct{}, handler *EventHandler) bool {
	handlerAddr := reflect.ValueOf(handler).Pointer()
	for onceHandler := range onceHandlers {
		if reflect.ValueOf(onceHandler).Pointer() == handlerAddr {
			return true
		}
	}
	return false
}

// RemoveHandler 移除特定处理器
func (d *defaultDispatcher) RemoveHandler(eventName string, handler EventHandler) error {
	if eventName == "" || handler == nil {
		return fmt.Errorf("invalid event name or handler")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	handlers, exists := d.handlers[eventName]
	if !exists {
		return fmt.Errorf("event %s not found", eventName)
	}

	// 过滤掉需要移除的处理器
	updatedHandlers := handlers[:0]
	for _, h := range handlers {
		if &h != &handler {
			updatedHandlers = append(updatedHandlers, h)
		}
	}
	d.handlers[eventName] = updatedHandlers

	return nil
}

// ClearHandlers 清除某事件的所有处理器
func (d *defaultDispatcher) ClearHandlers(eventName string) {
	if eventName == "" {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.handlers, eventName)
	delete(d.once, eventName)
}
