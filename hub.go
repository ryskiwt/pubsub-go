package pubsub

import (
	"context"
	"sync"

	"github.com/gobwas/glob"
)

// Hub represents PubSubHub.
type Hub interface {
	// Pub publishes message.
	Pub(topic string, msg interface{})
	// Sub adds subscription and returns channel to subscribe.
	Sub(topic string) <-chan interface{}
	// Unsub removes subscription corresponding to the submitted channel.
	Unsub(topic string, ch <-chan interface{})
	// PSub adds pattern subscriptions and returns channel to subscribe.
	PSub(pattern string) <-chan interface{}
	// PUnsub removes pattern subscription corresponding to the submitted channel.
	PUnsub(pattern string, ch <-chan interface{})
	// Close closes Hub.
	Close()
	// Context returns context.
	Context() context.Context
}

// NewHub creates a new Hub.
func NewHub(queueSize int) Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &hub{
		queueSize: queueSize,
		topics:    make(map[string]Topic),
		ptopics:   make(map[string]Topic),
		regexps:   make(map[string]glob.Glob),
		ctx:       ctx,
		cancel:    cancel,
	}
}

type hub struct {
	queueSize int
	topics    map[string]Topic
	ptopics   map[string]Topic
	regexps   map[string]glob.Glob
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func (h *hub) Pub(topic string, msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// sub
	t, ok := h.topics[topic]
	if !ok {
		return
	}
	_ = t.Pub(msg)

	// psub
	for pattern, t := range h.ptopics {
		if h.regexps[pattern].Match(topic) {
			_ = t.Pub(msg)
		}
	}
}

func (h *hub) Sub(topic string) <-chan interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	t, ok := h.topics[topic]
	if !ok {
		t = NewTopic(h.ctx, h.queueSize)
		h.topics[topic] = t
	}

	ch, _ := t.Sub()
	return ch
}

func (h *hub) Unsub(topic string, ch <-chan interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	t, ok := h.topics[topic]
	if !ok {
		return
	}
	_ = t.Unsub(ch)

	if t.SubLen() == 0 {
		t.Close()
		delete(h.topics, topic)
	}
}

func (h *hub) PSub(pattern string) <-chan interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	t, ok := h.ptopics[pattern]
	if !ok {
		t = NewTopic(h.ctx, h.queueSize)
		h.ptopics[pattern] = t
		h.regexps[pattern] = glob.MustCompile(pattern)
	}

	ch, _ := t.Sub()
	return ch
}

func (h *hub) PUnsub(pattern string, ch <-chan interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	t, ok := h.ptopics[pattern]
	if !ok {
		return
	}
	_ = t.Unsub(ch)

	if t.SubLen() == 0 {
		t.Close()
		delete(h.ptopics, pattern)
	}
}

func (h *hub) Close() {
	h.cancel()
}

func (h *hub) Context() context.Context {
	return h.ctx
}
