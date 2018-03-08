package pubsub

import (
	"context"
	"errors"
	"sync"
)

// ErrTopicAlreadyClosed returns when the topic is already closed.
var ErrTopicAlreadyClosed = errors.New("topic is already closed")

// Topic represents message publish/subscribe interface for one topic.
type Topic interface {
	// Pub publishes message.
	Pub(msg interface{}) error
	// Sub adds subscription and returns channel to subscribe.
	Sub() (<-chan interface{}, error)
	// Unsub removes subscription corresponding to the submitted channel.
	Unsub(ch <-chan interface{}) error
	// SubLen returns length of subscribers.
	SubLen() int
	// Close closes Topic.
	Close()
	// Context returns context.
	Context() context.Context
}

// NewTopic creates a new Topic instance.
func NewTopic(ctx context.Context, queueSize int) Topic {
	ctx, cancel := context.WithCancel(ctx)
	t := topic{
		pubChan:  make(chan interface{}, queueSize),
		subChans: make(map[<-chan interface{}]chan interface{}),
		ctx:      ctx,
		cancel:   cancel,
	}

	go t.run()
	return &t
}

type topic struct {
	count    int64
	pubChan  chan interface{}
	subChans map[<-chan interface{}]chan interface{}
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func (t *topic) run() {
	go func() {
		<-t.ctx.Done()
		close(t.pubChan)
		t.unsubAll()
	}()

	for msg := range t.pubChan {
		t.pub(msg)
	}
}

func (t *topic) Pub(msg interface{}) error {
	select {
	case <-t.ctx.Done():
		return ErrTopicAlreadyClosed
	default:
	}

	t.pubChan <- msg
	return nil
}

func (t *topic) pub(msg interface{}) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, subChan := range t.subChans {
		subChan <- msg
	}
}

func (t *topic) Sub() (<-chan interface{}, error) {
	select {
	case <-t.ctx.Done():
		return nil, ErrTopicAlreadyClosed
	default:
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan interface{}, cap(t.pubChan))
	t.subChans[ch] = ch
	return ch, nil
}

func (t *topic) unsub(ch <-chan interface{}) {
	close(t.subChans[ch])
	delete(t.subChans, ch)
}

func (t *topic) Unsub(ch <-chan interface{}) error {
	select {
	case <-t.ctx.Done():
		return ErrTopicAlreadyClosed
	default:
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.unsub(ch)
	return nil
}

func (t *topic) unsubAll() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for ch := range t.subChans {
		t.unsub(ch)
	}
	t.subChans = nil
}

func (t *topic) SubLen() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.subChans)
}

func (t *topic) Close() {
	t.cancel()
}

func (t *topic) Context() context.Context {
	return t.ctx
}
