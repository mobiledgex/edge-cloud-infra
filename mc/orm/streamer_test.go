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

	// Create another subscriber, who unsubscribes in between
	wg.Add(1)
	go func() {
		defer wg.Done()
		id := 5
		streamCh := streamer.Subscribe()
		rcvdMsgs := []int{}
		for streamMsg := range streamCh {
			rcvdMsg := streamMsg.(int)
			if rcvdMsg == 5 {
				streamer.Unsubscribe(streamCh)
				break
			}
			rcvdMsgs = append(rcvdMsgs, rcvdMsg)
		}
		require.NotEqual(t, sendMsgs, rcvdMsgs, fmt.Sprintf("Client %d: shoud not receive all the msgs", id))
		require.Equal(t, 5, len(rcvdMsgs), fmt.Sprintf("Client %d: match received msgs", id))
	}()

	// Wait for all subscribers to finish reading streamed messages
	wg.Wait()
}

type TestKey struct {
	id   int
	name string
}

var streamTest = &StreamObj{}

func TestStreamMaps(t *testing.T) {
	var wg sync.WaitGroup
	testKey1 := TestKey{id: 1, name: "testKey1"}
	testKey2 := TestKey{id: 2, name: "testKey2"}
	// Msgs to be published to all subscribers
	sendMsgs := []int{}
	for id := 0; id < 10; id++ {
		sendMsgs = append(sendMsgs, id)
	}

	// Start publishing messages
	go func() {
		streamer := NewStreamer()
		defer streamer.Stop()
		err := streamTest.Add(testKey1, streamer)
		require.Nil(t, err, "successfully added stream")
		for _, msg := range sendMsgs {
			streamer.Publish(msg)
			time.Sleep(2 * time.Millisecond)
		}
		streamTest.Remove(testKey1)
	}()

	// Subscriber
	subscriberFunc := func(wg *sync.WaitGroup, key interface{}, id int) {
		defer wg.Done()
		streamer := streamTest.Get(key)
		require.NotNil(t, streamer, "stream exists")
		streamCh := streamer.Subscribe()
		rcvdMsgs := []int{}
		for streamMsg := range streamCh {
			rcvdMsgs = append(rcvdMsgs, streamMsg.(int))
		}
		require.Equal(t, sendMsgs, rcvdMsgs, fmt.Sprintf("Client %d: match received msgs", id))
	}

	// Create multiple subscribers started with some time gap
	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(i) * time.Millisecond)
		wg.Add(1)
		go subscriberFunc(&wg, testKey1, i)
	}

	// Start another publisher for same key
	go func() {
		streamer := NewStreamer()
		defer streamer.Stop()
		err := streamTest.Add(testKey1, streamer)
		require.NotNil(t, err, "Publisher is busy")
	}()

	// Start another publisher for different key
	go func() {
		streamer := NewStreamer()
		defer streamer.Stop()
		err := streamTest.Add(testKey2, streamer)
		require.Nil(t, err, "successfully added stream")
		for _, msg := range sendMsgs {
			streamer.Publish(msg)
			time.Sleep(1 * time.Millisecond)
		}
		streamTest.Remove(testKey2)
	}()

	// Create multiple subscribers started with some time gap
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i) * time.Millisecond)
		wg.Add(1)
		go subscriberFunc(&wg, testKey2, i)
	}

	// Wait for all subscribers to finish reading streamed messages
	wg.Wait()
}
