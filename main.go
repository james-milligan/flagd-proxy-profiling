package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/james-milligan/flagd-proxy-profiling/pkg/config"
	"github.com/james-milligan/flagd-proxy-profiling/pkg/handler"
	itrigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger"
	trigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger/file"
)

func main() {
	configFilepath := ""
	if len(os.Args) == 2 {
		configFilepath = os.Args[1]
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	cfg, err := config.NewConfig(configFilepath)
	if err != nil {
		log.Fatal(err)
	}
	var trg itrigger.Trigger
	switch cfg.TriggerType {
	case config.FilepathTrigger:
		trg = trigger.NewFilePathTrigger(cfg.FileTriggerConfig)
	default:
		log.Fatalf("unrecognized trigger type %s", cfg.TriggerType)
	}

	h := handler.NewHandler(cfg.HandlerConfig, trg)
	results, err := h.Profile(ctx, cfg.Tests)
	fmt.Println(err, results)
}
