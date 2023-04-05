package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	trigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger/file"
	"github.com/james-milligan/flagd-proxy-profiling/pkg/watcher"
)

const (
	host            = "localhost"
	port     uint16 = 8080
	routines        = 130
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	readyWg := sync.WaitGroup{}
	readyWg.Add(routines)

	finishedWg := sync.WaitGroup{}
	finishedWg.Add(routines)

	if err := trigger.SetupFile(); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < routines; i++ {
		go func() {
			w := watcher.NewWatcher(host, port)
			go func() {
				if err := w.StartWatcher(ctx); err != nil {
					log.Fatal(err)
				}
			}()
			<-w.Ready
			readyWg.Done()
			if err := w.Wait(); err != nil {
				log.Fatal(err)
			}
			finishedWg.Done()
		}()
	}

	fmt.Printf("starting %d watchers...\n", routines)
	readyWg.Wait() // all routines are ready
	fmt.Println("all watchers ready, starting timer and writing to file...")
	start := time.Now()

	if err := trigger.UpdateFile(); err != nil {
		log.Fatal(err)
	}

	finishedWg.Wait()

	end := time.Now()

	fmt.Println("process took", end.Sub(start))

}
