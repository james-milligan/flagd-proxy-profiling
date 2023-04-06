package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/james-milligan/flagd-proxy-profiling/pkg/handler"
	trigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger/file"
)

const (
	host             = "localhost"
	port      uint16 = 8080
	watchers         = 10000
	filepath         = "/Users/jamesmilligan/code/flagd-1/config/samples/example_flags.json"
	startFile        = "./config/start-spec.json"
	endFile          = "./config/end-spec.json"
	outFile          = "./profiling-results.json"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	trigger := trigger.NewFilePathTrigger(trigger.FilePathTriggerConfig{
		StartFile:  startFile,
		EndFile:    endFile,
		TargetFile: filepath,
	})
	h := handler.NewHandler(handler.HandlerConfig{
		Host:     host,
		Port:     port,
		FilePath: filepath,
		OutFile:  outFile,
	}, trigger)
	results, err := h.Profile(ctx, handler.TestConfig{
		Watchers: watchers,
		Repeats:  5,
		Delay:    time.Second * 1,
	})
	fmt.Println(err, results)
}
