package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/stretchr/testify/require"
)

func startMain(t *testing.T) (chan struct{}, error) {
	mainStarted = make(chan struct{})
	mainDone := make(chan struct{})
	go func() {
		main()
		close(mainDone)
	}()
	// wait until main is ready
	select {
	case <-mainStarted:
	case <-mainDone:
		return nil, fmt.Errorf("main unexpectedly quit")
	}
	return mainDone, nil
}

func TestFRM(t *testing.T) {
	var err error
	log.SetDebugLevel(log.DebugLevelApi | log.DebugLevelNotify | log.DebugLevelInfra)
	log.InitTracer(nil)
	defer log.FinishTracer()
	log.StartTestSpan(context.Background())

	// FRM is driven by controller
	ctrlHandler := notify.NewDummyHandler()
	ctrlMgr := notify.ServerMgr{}
	ctrlHandler.RegisterServer(&ctrlMgr)
	notifyAddr := "127.0.0.1:61245"
	ctrlMgr.Start("ctrl", notifyAddr, nil)

	os.Args = append(os.Args, "-notifyAddrs")
	os.Args = append(os.Args, notifyAddr)
	mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		require.Nil(t, err, "start main")
		return
	}
	defer func() {
		// closing the signal channel triggers main to exit
		close(sigChan)
		// wait until main is done so it can clean up properly
		<-mainDone
		ctrlMgr.Stop()
	}()

	notifyClient.WaitForConnect(1)
	stats := notify.Stats{}
	notifyClient.GetStats(&stats)
	require.Equal(t, uint64(1), stats.Connects)
}
