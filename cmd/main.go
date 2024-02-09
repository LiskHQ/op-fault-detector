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
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/api"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/faultdetector"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/LiskHQ/op-fault-detector/pkg/utils/notification"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

// App encapsulates start and stop logic for the whole application.
type App struct {
	ctx           context.Context
	logger        log.Logger
	errChan       chan error
	config        *config.Config
	wg            *sync.WaitGroup
	apiServer     *api.HTTPServer
	faultDetector *faultdetector.FaultDetector
	notification  *notification.Channel
}

// NewApp returns [App] with all the initialized services and variables.
func NewApp(ctx context.Context, logger log.Logger) (*App, error) {
	reg := prometheus.NewRegistry()
	// Adding Go process related metric
	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	configFilepath := flag.String("config", "./config.yaml", "Path to the config file")
	flag.Parse()
	config, err := getAppConfig(logger, *configFilepath)
	if err != nil {
		logger.Errorf("Failed at parsing config with error %w", err)
		return nil, err
	}

	wg := sync.WaitGroup{}
	errorChan := make(chan error, 1)

	// Initialize notification service
	var notificationChannel *notification.Channel
	if config.Notification.Enable {
		notificationChannel, err = notification.NewChannel(
			ctx,
			logger,
			config.Notification,
		)
		if err != nil {
			logger.Errorf("Failed to initialize notification service, error: %v.", err)
			return nil, err
		}
	}

	// Start Fault Detector
	faultDetector, err := faultdetector.NewFaultDetector(
		ctx,
		logger,
		errorChan,
		&wg,
		config.FaultDetectorConfig,
		reg,
		notificationChannel,
	)

	if err != nil {
		logger.Errorf("Failed to create fault detector service.")
		return nil, err
	}

	// Start API Server
	apiServer := api.NewHTTPServer(ctx, logger, &wg, config, errorChan)

	// Register metrics endpoint on the apiServer
	apiServer.RegisterHandler("GET", "/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg, ProcessStartTime: time.Now()}))

	return &App{
		ctx:           ctx,
		logger:        logger,
		errChan:       errorChan,
		config:        config,
		wg:            &wg,
		apiServer:     apiServer,
		faultDetector: faultDetector,
		notification:  notificationChannel,
	}, nil
}

// This will start the application by starting API Server and Fault Detector services.
func (app *App) Start() {
	doneChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	app.wg.Add(1)
	go app.faultDetector.Start()

	app.wg.Add(1)
	go app.apiServer.Start()

	go func() {
		app.wg.Wait()
		close(doneChan)
	}()

	for {
		select {
		case <-doneChan:
			app.stop()
			return

		case <-signalChan:
			app.stop()
			return

		case err := <-app.errChan:
			app.logger.Errorf("Received error of %v", err)
			if app.notification != nil {
				if err := app.notification.Notify("Error while starting application"); err != nil {
					app.logger.Errorf("Error while sending notification, error: %w", err)
				}
			}
			return
		}
	}
}

func (app *App) stop() {
	app.faultDetector.Stop()
	err := app.apiServer.Stop()
	if err != nil {
		app.logger.Error("Server shutdown not successful: %w", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := log.NewDefaultProductionLogger()
	if err != nil {
		logger.Errorf("Failed to create logger, %w", err)
		return
	}

	app, err := NewApp(ctx, logger)
	if err != nil {
		logger.Errorf("Failed to create app, %w", err)
		return
	}

	logger.Infof("Starting app...")
	app.Start()
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
