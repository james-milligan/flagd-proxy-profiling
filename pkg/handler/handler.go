package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"buf.build/gen/go/open-feature/flagd/grpc/go/sync/v1/syncv1grpc"
	"github.com/james-milligan/flagd-proxy-profiling/pkg/client"
	trigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger/file"
	"github.com/james-milligan/flagd-proxy-profiling/pkg/watcher"
)

type Handler struct {
	config HandlerConfig
}

type HandlerConfig struct {
	FilePath string
	Host     string
	Port     uint16
	OutFile  string
}

type TestConfig struct {
	Listeners int
	Repeats   int
	Delay     time.Duration
}

type TestResult struct {
	TotalTime      time.Duration `json:"totalTime"`
	TimePerWatcher time.Duration `json:"timePerWatcher"`
}

type ProfilingResults struct {
	Tests                 []TestResult  `json:"tests"`
	AverageTotalDuration  time.Duration `json:"averageTotalDuration"`
	AverageTimePerWatcher time.Duration `json:"averageTimePerWatcher"`
	Listeners             int           `json:"listeners"`
	Repeats               int           `json:"repeats"`
}

func (h *Handler) writeFile(results ProfilingResults) error {
	resB, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.config.OutFile, resB, 0644)
}

func NewHandler(config HandlerConfig) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) Profile(ctx context.Context, config TestConfig) (ProfilingResults, error) {
	results := []TestResult{}
	for i := 1; i <= config.Repeats; i++ {
		fmt.Printf("starting profile %d\n", i)
		res := h.runTest(ctx, config.Listeners)
		results = append(results, res)
		fmt.Println("-----------------------")
		time.Sleep(config.Delay)
	}
	timePer := time.Duration(0)
	totalTime := time.Duration(0)
	for _, res := range results {
		timePer += res.TimePerWatcher
		totalTime += res.TotalTime
	}
	out := ProfilingResults{
		Listeners:             config.Listeners,
		Repeats:               config.Repeats,
		Tests:                 results,
		AverageTotalDuration:  totalTime / time.Duration(config.Repeats),
		AverageTimePerWatcher: timePer / time.Duration(config.Repeats),
	}
	return out, h.writeFile(out)
}

func (h *Handler) runTest(ctx context.Context, listeners int) TestResult {
	readyWg := sync.WaitGroup{}
	readyWg.Add(listeners)

	finishedWg := sync.WaitGroup{}
	finishedWg.Add(listeners)

	fmt.Printf("starting %d watchers...\n", listeners)
	var c syncv1grpc.FlagSyncServiceClient
	var err error
	for i := 0; i < listeners; i++ {
		if i%250 == 0 {
			c, err = client.NewClient(client.ClientConfig{
				Host: h.config.Host,
				Port: h.config.Port,
			})
			if err != nil {
				log.Fatal(err)
			}
		}
		go func() {
			w := watcher.NewWatcher(c)
			go func() {
				if err := w.StartWatcher(ctx); err != nil {
					fmt.Println(err)
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

	fmt.Println("waiting for watchers to be ready...")
	readyWg.Wait() // all routines are ready
	fmt.Println("all watchers ready, starting timer and writing to file...")
	start := time.Now()

	if err := trigger.UpdateFile(h.config.FilePath); err != nil {
		log.Fatal(err)
	}

	finishedWg.Wait()

	end := time.Now()
	fmt.Println("done")

	timeTaken := end.Sub(start)
	fmt.Println("process took", timeTaken)
	fmt.Println("time per run", timeTaken/time.Duration(listeners))
	return TestResult{
		TotalTime:      timeTaken,
		TimePerWatcher: timeTaken / time.Duration(listeners),
	}
}
