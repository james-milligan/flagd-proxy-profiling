package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/james-milligan/flagd-proxy-profiling/pkg/handler"
)

const (
	host             = "localhost"
	port      uint16 = 8080
	listeners        = 10000
	filepath         = "/Users/jamesmilligan/code/flagd-1/config/samples/example_flags.json"
	outFile          = "./profiling-results.json"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	h := handler.NewHandler(handler.HandlerConfig{
		Host:     host,
		Port:     port,
		FilePath: filepath,
		OutFile:  outFile,
	})
	results, err := h.Profile(ctx, handler.TestConfig{
		Listeners: listeners,
		Repeats:   5,
		Delay:     time.Second * 1,
	})
	fmt.Println(err, results)
}
