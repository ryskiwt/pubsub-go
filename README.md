## pubsub-go

simple pub/sub library written in golang

## features
- quite simple interface (Pub/Sub/PSub)
- type of topic is string
- type of message is interface{}
- support pattern subscribe (PSub/PUnsub)

## interface

```go
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
```

## example

```go
package main

import (
	"fmt"
	"time"

	pubsub "github.com/ryskiwt/pubsub-go"
)

func main() {

	h := pubsub.NewHub(32)

	subChan11 := h.Sub("/some/topic/1")
	go func() {
		for msg := range subChan11 {
			fmt.Printf("subChan11, TOPIC: /some/topic/1, MSG: %s\n", msg)
		}
	}()
	subChan12 := h.Sub("/some/topic/1")
	go func() {
		for msg := range subChan12 {
			fmt.Printf("subChan12, TOPIC: /some/topic/1, MSG: %s\n", msg)
		}
	}()

	subChan2 := h.Sub("/some/topic/2")
	go func() {
		for msg := range subChan2 {
			fmt.Printf("subChan2,  TOPIC: /some/topic/2, MSG: %s\n", msg)
		}
	}()

	subChan3 := h.PSub("/some/topic/*")
	go func() {
		for msg := range subChan3 {
			fmt.Printf("subChan3,  TOPIC: /some/topic/*, MSG: %s\n", msg)
		}
	}()

	h.Pub("/some/topic/1", "message 1 !")
	h.Pub("/some/topic/2", "message 2 !")
	h.Pub("/some/topic/3", "message 3 !")
	h.Pub("/some/topic/1", "message 4 !")
	h.Pub("/some/topic/2", "message 5 !")
	h.Pub("/some/topic/3", "message 6 !")

	<-time.After(time.Second)
	h.Unsub("/some_topic/1", subChan11)
	h.Unsub("/some_topic/1", subChan12)
	h.Unsub("/some_topic/2", subChan2)
	h.PUnsub("/some_topic/*", subChan3)
}
```

```sh
$ cd example
$ go run main.go
subChan11, TOPIC: /some/topic/1, MSG: message 1 !
subChan12, TOPIC: /some/topic/1, MSG: message 1 !
subChan12, TOPIC: /some/topic/1, MSG: message 4 !
subChan3,  TOPIC: /some/topic/*, MSG: message 1 !
subChan3,  TOPIC: /some/topic/*, MSG: message 2 !
subChan3,  TOPIC: /some/topic/*, MSG: message 4 !
subChan3,  TOPIC: /some/topic/*, MSG: message 5 !
subChan11, TOPIC: /some/topic/1, MSG: message 4 !
subChan2,  TOPIC: /some/topic/2, MSG: message 2 !
subChan2,  TOPIC: /some/topic/2, MSG: message 5 !
```
