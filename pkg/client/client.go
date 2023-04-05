package client

import (
	"fmt"
	"net/http"

	syncv1 "buf.build/gen/go/open-feature/flagd/bufbuild/connect-go/sync/v1/syncv1connect"
)

type ClientConfig struct {
	Host string
	Port uint16
}

func NewClient(config ClientConfig) syncv1.FlagSyncServiceClient {
	url := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	return syncv1.NewFlagSyncServiceClient(http.DefaultClient, url)
}
