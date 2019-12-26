package orm

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamer(t *testing.T) {
	var wg sync.WaitGroup

	streamer := NewStreamer()
	go streamer.Start()

	// Msgs to be published to all subscribers
	sendMsgs := []int{}
	for id := 0; id < 10; id++ {
		sendMsgs = append(sendMsgs, id)
	}

	// Subscriber
	subscriberFunc := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		streamCh := streamer.Subscribe()
		rcvdMsgs := []int{}
		for streamMsg := range streamCh {
			rcvdMsgs = append(rcvdMsgs, streamMsg.(int))
		}
		require.Equal(t, sendMsgs, rcvdMsgs, fmt.Sprintf("Client %d: match received msgs", id))
	}

	// Create multiple subscribers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go subscriberFunc(&wg, i)
	}

	// Start publishing messages
	go func() {
		for _, msg := range sendMsgs {
			streamer.Publish(msg)
			time.Sleep(1 * time.Millisecond)
		}
		streamer.Stop()
	}()

	// Create some more subscribers started after a while
	for i := 3; i < 5; i++ {
		time.Sleep(2 * time.Millisecond)
		wg.Add(1)
		go subscriberFunc(&wg, i)
	}

	// Wait for all subscribers to finish reading streamed messages
	wg.Wait()
}
