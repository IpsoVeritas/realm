package pubsub

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type MiniPubSub struct {
	url string
}

func NewMiniPubSub(url string) *MiniPubSub {
	return &MiniPubSub{
		url: url,
	}
}

func (p *MiniPubSub) Publish(topic string, data string) error {
	buf := bytes.NewBufferString(data)
	resp, err := http.Post(fmt.Sprintf("%s?topic=%s", p.url, topic), "test/plain", buf)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("Expected status 201, got %v", resp.StatusCode)
	}

	return nil
}

func (p *MiniPubSub) Subscribe(group string, topic string) (Subscriber, error) {
	return &MiniSubscriber{
		group: group,
		topic: topic,
	}, nil
}

func (p *MiniPubSub) DeleteTopic(topic string) error {
	return nil
}

func (p *MiniPubSub) Stop() {
	return
}

type MiniSubscriber struct {
	group string
	topic string
	url   string
	stop  chan struct{}
}

func (s *MiniSubscriber) Pull(timeout time.Duration) (string, int) {
	resp, err := http.Get(fmt.Sprintf("%s?topic=%s&group=%s", s.url, s.topic, s.group))
	if err != nil {
		return "", ERROR
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ERROR
	}

	return string(bytes), SUCCESS
}

func (s *MiniSubscriber) Stop(timeout time.Duration) {
	close(s.stop)
	return
}

func (s *MiniSubscriber) Chan() chan string {
	c := make(chan string)
	go func() {
		select {
		default:
			for {
				m, i := s.Pull(1 * time.Second)
				if i == SUCCESS {
					c <- m
				} else {
					time.Sleep(10 * time.Millisecond)
				}
			}
		case <-s.stop:
			return
		}
	}()

	return c
}
