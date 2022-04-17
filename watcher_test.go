package curator

import (
	"runtime"
	"sync"
	"testing"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
)

func TestWatchers(t *testing.T) {
	var events [3][]*zk.Event

	wg := sync.WaitGroup{}

	w := NewWatchers(NewWatcher(func(event *zk.Event) {
		events[0] = append(events[0], event)
		wg.Done()
	}))

	w1 := w.Add(NewWatcher(func(event *zk.Event) {
		events[1] = append(events[1], event)
		wg.Done()
	}))

	w2 := w.Add(NewWatcher(func(event *zk.Event) {
		events[2] = append(events[2], event)
		wg.Done()
	}))

	assert.Equal(t, w1, w.watchers[1])
	assert.Equal(t, w2, w.watchers[2])

	c := make(chan zk.Event)

	go w.Watch(c)

	evt := zk.Event{}

	wg.Add(3)
	c <- evt
	wg.Wait()

	close(c)

	assert.Equal(t, []*zk.Event{&evt}, events[0])
	assert.Equal(t, []*zk.Event{&evt}, events[1])
	assert.Equal(t, []*zk.Event{&evt}, events[2])

	// remove watcher and fire event again
	assert.Equal(t, w.Remove(w1), w1)
	assert.Equal(t, w.Remove(w2), w2)

	assert.Equal(t, 1, len(w.watchers))

	c = make(chan zk.Event)

	go w.Watch(c)

	evt = zk.Event{}

	wg.Add(1)
	c <- evt
	wg.Wait()

	runtime.Gosched()

	close(c)

	assert.Equal(t, 2, len(events[0]))
	assert.Equal(t, 1, len(events[1]))
	assert.Equal(t, 1, len(events[2]))
	assert.Equal(t, &evt, events[0][1])
}
