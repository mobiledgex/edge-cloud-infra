package orm

import (
	"sync"
)

type Streamer struct {
	buffer      []interface{}
	stopCh      chan struct{}
	publishCh   chan interface{}
	subscribeCh chan chan interface{}
	mux         sync.Mutex
}

func NewStreamer() *Streamer {
	return &Streamer{
		stopCh:      make(chan struct{}),
		publishCh:   make(chan interface{}, 1),
		subscribeCh: make(chan chan interface{}, 1),
	}
}

func (b *Streamer) Start() {
	subs := map[chan interface{}]struct{}{}
	for {
		b.mux.Lock()
		select {
		case <-b.stopCh:
			for msgCh := range subs {
				close(msgCh)
			}
			b.mux.Unlock()
			return
		case msgCh := <-b.subscribeCh:
			subs[msgCh] = struct{}{}
			for _, msg := range b.buffer {
				select {
				case msgCh <- msg:
				default:
				}
			}
		case msg := <-b.publishCh:
			// Buffer all the streamed messages till now,
			// so that a newly joined subscriber can get
			// complete list of messages
			b.buffer = append(b.buffer, msg)
			for msgCh := range subs {
				select {
				case msgCh <- msg:
				default:
				}
			}
		}
		b.mux.Unlock()
	}
}

func (b *Streamer) Stop() {
	close(b.stopCh)
}

func (b *Streamer) Subscribe() chan interface{} {
	msgCh := make(chan interface{}, 20)
	b.subscribeCh <- msgCh
	return msgCh
}

func (b *Streamer) Publish(msg interface{}) {
	b.publishCh <- msg
}
