package pubsub

import (
	"sync"
	"time"
)

type InmemPubSub struct {
	subscribers map[string]map[string]chan string
	mu          sync.Mutex
}

func NewInmemPubSub() *InmemPubSub {
	i := &InmemPubSub{
		subscribers: make(map[string]map[string]chan string),
	}

	return i
}

func (i *InmemPubSub) Publish(topic string, data string) error {

	_, ok := i.subscribers[topic]
	if !ok {
		return nil // nobody is listening so let's just return...
	}

	for _, c := range i.subscribers[topic] {
		select {
		case c <- data:
		case <-time.After(time.Second * 1):
		}
	}

	return nil
}

func (i *InmemPubSub) Subscribe(group string, topic string) (Subscriber, error) {

	_, ok := i.subscribers[topic]
	if !ok {
		i.mu.Lock()
		i.subscribers[topic] = make(map[string]chan string)
		i.mu.Unlock()
	}
	_, ok = i.subscribers[topic][group]
	if !ok {
		i.mu.Lock()
		i.subscribers[topic][group] = make(chan string, 10000)
		i.mu.Unlock()
	}

	sub := &InmemSubscriber{
		i:     i,
		group: group,
		topic: topic,
	}

	return sub, nil
}

func (i *InmemPubSub) DeleteTopic(topic string) error {

	_, ok := i.subscribers[topic]
	if ok {
		delete(i.subscribers, topic)
	}

	return nil
}

func (i *InmemPubSub) Stop() {

	return
}

type InmemSubscriber struct {
	i     *InmemPubSub
	group string
	topic string
}

func (s *InmemSubscriber) Pull(timeout time.Duration) (string, int) {

	_, ok := s.i.subscribers[s.topic]
	if ok {
		_, ok = s.i.subscribers[s.topic][s.group]
		if ok {
			select {
			case msg := <-s.i.subscribers[s.topic][s.group]:
				return msg, SUCCESS
			case <-time.After(time.Second * timeout):
				return "", TIMEOUT
			}
		}
	}

	return "", ERROR
}

func (s *InmemSubscriber) Chan() chan string {
	return s.i.subscribers[s.topic][s.group]
}

func (s *InmemSubscriber) Stop(timeout time.Duration) {
	return
}
