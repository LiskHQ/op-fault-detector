// CLI to run fault detector service
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"

	"github.com/LiskHQ/op-fault-detector/pkg/api"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/spf13/viper"
)

var apiServer *api.HTTPServer

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := log.NewDefaultProductionLogger()
	if err != nil {
		panic(err)
	}

	configFilepath := flag.String("config", "./config.yaml", "Path to the config file")
	flag.Parse()
	config, err := getAppConfig(logger, *configFilepath)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	doneChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start Fault Detector

	// Start API Server
	serverChan := make(chan error, 1)
	apiServer = api.NewHTTPServer(ctx, logger, &wg, config, serverChan)
	wg.Add(1)
	go apiServer.Start()

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	for {
		select {
		case <-doneChan:
			performCleanup(logger)
			return

		case <-signalChan:
			performCleanup(logger)
			return

		case err := <-serverChan:
			logger.Errorf("Received error of %v", err)
			return
		}
	}
}

// getAppConfig is the function that takes in the absolute path to the config file, parses the content and returns it.
func getAppConfig(logger log.Logger, configFilepath string) (*config.Config, error) {
	configDir := path.Dir(configFilepath)
	configFilenameWithExt := path.Base(configFilepath)

	splits := strings.FieldsFunc(configFilenameWithExt, func(r rune) bool { return r == '.' })
	configType := splits[len(splits)-1] // Config file extension

	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("$HOME/.op-fault-detector")
	viper.AddConfigPath(configDir)
	viper.SetConfigName(configFilenameWithExt)
	viper.SetConfigType(configType)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load the config from disk: %w", err)
	}

	var config config.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.New("failed to unmarshal config. Verify the 'Config' struct definition in 'pkg/config/config.go'")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func performCleanup(logger log.Logger) {
	err := apiServer.Stop()
	if err != nil {
		logger.Error("Server shutdown not successful: %w", err)
	}
}
