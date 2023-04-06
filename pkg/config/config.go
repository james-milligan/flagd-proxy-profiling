package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/james-milligan/flagd-proxy-profiling/pkg/handler"
	trigger "github.com/james-milligan/flagd-proxy-profiling/pkg/trigger/file"
)

type TriggerType string

const (
	FilepathTrigger TriggerType = "filepath"

	defaultHost                    = "localhost"
	defaultPort             uint16 = 8080
	defaultStartFile               = "./config/start-spec.json"
	defaultEndFile                 = "./config/end-spec.json"
	defaultTargetFile              = "./target"
	defaultTargetFileSource        = "./target"
	defaultOutTarget               = "./profiling-results.json"
)

var defaultTests = []handler.TestConfig{
	{
		Watchers: 1,
		Repeats:  5,
		Delay:    time.Second * 1,
	},
	{
		Watchers: 10,
		Repeats:  5,
		Delay:    time.Second * 1,
	},
	{
		Watchers: 100,
		Repeats:  5,
		Delay:    time.Second * 1,
	},
	{
		Watchers: 1000,
		Repeats:  5,
		Delay:    time.Second * 1,
	},
	{
		Watchers: 10000,
		Repeats:  5,
		Delay:    time.Second * 1,
	},
}

type Config struct {
	TriggerType       TriggerType                   `json:"triggerType"`
	FileTriggerConfig trigger.FilePathTriggerConfig `json:"fileTriggerConfig"`
	HandlerConfig     handler.HandlerConfig         `json:"handlerConfig"`
	Tests             []handler.TestConfig
}

func NewConfig(filepath string) (*Config, error) {
	config := &Config{
		TriggerType: FilepathTrigger,
		FileTriggerConfig: trigger.FilePathTriggerConfig{
			StartFile:  defaultStartFile,
			EndFile:    defaultEndFile,
			TargetFile: defaultTargetFile,
		},
		HandlerConfig: handler.HandlerConfig{
			FilePath: defaultTargetFileSource,
			Host:     defaultHost,
			Port:     defaultPort,
			OutFile:  defaultOutTarget,
		},
		Tests: defaultTests,
	}
	if filepath != "" {
		b, err := os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, config); err != nil {
			return nil, err
		}
	}
	return config, nil
}
