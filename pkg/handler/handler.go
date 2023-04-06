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
	trigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger"
	"github.com/james-milligan/flagd-proxy-profiling/pkg/watcher"
)

type Handler struct {
	config  HandlerConfig
	trigger trigger.Trigger
}

type HandlerConfig struct {
	FilePath string
	Host     string
	Port     uint16
	OutFile  string
}

type TestConfig struct {
	Watchers int
	Repeats  int
	Delay    time.Duration
}

type TestResult struct {
	TotalTime      time.Duration `json:"totalTime"`
	TimePerWatcher time.Duration `json:"timePerWatcher"`
}

type ProfilingResults struct {
	Tests                 []TestResult  `json:"tests"`
	AverageTotalDuration  time.Duration `json:"averageTotalDuration"`
	AverageTimePerWatcher time.Duration `json:"averageTimePerWatcher"`
	Watchers              int           `json:"watchers"`
	Repeats               int           `json:"repeats"`
}

func NewHandler(config HandlerConfig, trigger trigger.Trigger) *Handler {
	return &Handler{
		config:  config,
		trigger: trigger,
	}
}

func (h *Handler) Profile(ctx context.Context, config TestConfig) (ProfilingResults, error) {
	if err := h.trigger.Setup(); err != nil {
		return ProfilingResults{}, err
	}
	results := []TestResult{}
	for i := 1; i <= config.Repeats; i++ {
		fmt.Printf("starting profile %d\n", i)
		res := h.runTest(ctx, config.Watchers)
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
		Watchers:              config.Watchers,
		Repeats:               config.Repeats,
		Tests:                 results,
		AverageTotalDuration:  totalTime / time.Duration(config.Repeats),
		AverageTimePerWatcher: timePer / time.Duration(config.Repeats),
	}
	return out, h.writeFile(out)
}

func (h *Handler) runTest(ctx context.Context, watchers int) TestResult {
	readyWg := sync.WaitGroup{}
	readyWg.Add(watchers)

	finishedWg := sync.WaitGroup{}
	finishedWg.Add(watchers)

	fmt.Printf("starting %d watchers...\n", watchers)
	var c syncv1grpc.FlagSyncServiceClient
	var err error
	for i := 0; i < watchers; i++ {
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

	if err := h.trigger.Update(); err != nil {
		log.Fatal(err)
	}

	finishedWg.Wait()

	end := time.Now()
	fmt.Println("done")

	timeTaken := end.Sub(start)
	fmt.Println("process took", timeTaken)
	fmt.Println("time per run", timeTaken/time.Duration(watchers))
	return TestResult{
		TotalTime:      timeTaken,
		TimePerWatcher: timeTaken / time.Duration(watchers),
	}
}

func (h *Handler) writeFile(results ProfilingResults) error {
	resB, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.config.OutFile, resB, 0644)
}
