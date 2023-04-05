package watcher

import (
	"context"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/open-feature/flagd/bufbuild/connect-go/sync/v1/syncv1connect"
	syncv1Types "buf.build/gen/go/open-feature/flagd/protocolbuffers/go/sync/v1"

	"github.com/bufbuild/connect-go"
	"github.com/james-milligan/flagd-proxy-profiling/pkg/client"
)

const (
	timeoutSeconds = 20
)

type Watcher struct {
	client syncv1.FlagSyncServiceClient
	Stream chan syncv1Types.SyncState
	Ready  chan struct{}
}

func NewWatcher(host string, port uint16) *Watcher {
	return &Watcher{
		Stream: make(chan syncv1Types.SyncState, 1),
		client: client.NewClient(client.ClientConfig{
			Host: host,
			Port: port,
		}),
		Ready: make(chan struct{}),
	}
}

func (w *Watcher) StartWatcher(ctx context.Context) error {
	res, err := w.client.SyncFlags(ctx, connect.NewRequest(&syncv1Types.SyncFlagsRequest{
		Selector: "file:./config/samples/example_flags.json",
	}))
	if err != nil {
		return err
	}
	ready := false
	for res.Receive() {
		w.Stream <- res.Msg().State
		if !ready {
			ready = true
			close(w.Ready)
		}
	}
	return res.Err()
}

func (w *Watcher) Wait() error {
	w.drainChan()
	select {
	case <-time.After(timeoutSeconds * time.Second):
		return fmt.Errorf("timeout out after %d", timeoutSeconds)
	case <-w.Stream:
		return nil
	}
}

func (w *Watcher) drainChan() {
	for {
		select {
		case <-w.Stream:
		default:
			return
		}
	}
}
