package main

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/api"
	v1 "github.com/LiskHQ/op-fault-detector/pkg/api/handlers/v1"
	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/faultdetector"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/LiskHQ/op-fault-detector/pkg/utils/notification"
	slack "github.com/LiskHQ/op-fault-detector/pkg/utils/notification/channel"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promClient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	slackClient "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	host                       = "127.0.0.1"
	port                       = 8088
	faultProofWindow           = 1000
	currentOutputIndex         = 1
	l1RpcApi                   = "https://rpc.notadegen.com/eth"
	l2RpcApi                   = "https://mainnet.optimism.io/"
	faultDetectorStateMismatch = "fault_detector_is_state_mismatch"
)

type parseMetric map[string]map[string]interface{}

type mockContractOracleAccessor struct {
	mock.Mock
}

type mockSlackClient struct {
	mock.Mock
}

func (o *mockContractOracleAccessor) GetNextOutputIndex() (*big.Int, error) {
	called := o.MethodCalled("GetNextOutputIndex")
	return called.Get(0).(*big.Int), called.Error(1)
}

func (o *mockContractOracleAccessor) GetL2Output(index *big.Int) (chain.L2Output, error) {
	called := o.MethodCalled("GetL2Output", index)
	return called.Get(0).(chain.L2Output), called.Error(1)
}

func (o *mockContractOracleAccessor) FinalizationPeriodSeconds() (*big.Int, error) {
	called := o.MethodCalled("FinalizationPeriodSeconds")
	return called.Get(0).(*big.Int), called.Error(1)
}

func (o *mockSlackClient) PostMessageContext(ctx context.Context, channelID string, options ...slackClient.MsgOption) (string, string, error) {
	called := o.MethodCalled("PostMessageContext")
	return called.Get(0).(string), called.Get(1).(string), called.Error(2)
}

var slackNotificationClient *mockSlackClient = new(mockSlackClient)

func randHash() (out common.Hash) {
	_, _ = crand.Read(out[:])
	return out
}

func prepareHTTPServer(t *testing.T, ctx context.Context, logger log.Logger, wg *sync.WaitGroup, erroChan chan error) *api.HTTPServer {
	router := gin.Default()
	return &api.HTTPServer{
		Server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", host, port),
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
		},
		Router:    router,
		Ctx:       ctx,
		Logger:    logger,
		Wg:        wg,
		ErrorChan: erroChan,
	}
}

func prepareNotification(t *testing.T, ctx context.Context, logger log.Logger) *notification.Notification {
	slackNotificationClient.On("PostMessageContext").Return("TestChannelID", "1234569.1000", nil)

	return &notification.Notification{
		Slack: &slack.Slack{
			Client:    slackNotificationClient,
			ChannelID: "string",
			Ctx:       ctx,
			Logger:    logger,
		},
	}
}

func prepareFaultDetector(t *testing.T, ctx context.Context, logger log.Logger, wg *sync.WaitGroup, reg *prometheus.Registry, config *config.Config, erroChan chan error, mock bool) *faultdetector.FaultDetector {
	var fd *faultdetector.FaultDetector
	if !mock {
		fd, _ = faultdetector.NewFaultDetector(ctx, logger, erroChan, wg, config.FaultDetectorConfig, reg, &notification.Notification{})
	} else {
		metrics := faultdetector.NewFaultDetectorMetrics(reg)

		// Create chain API clients
		l1RpcApi, _ := chain.GetAPIClient(ctx, l1RpcApi, logger)
		l2RpcApi, _ := chain.GetAPIClient(ctx, l2RpcApi, logger)

		latestL2BlockNumber, _ := l2RpcApi.GetLatestBlockNumber(ctx)

		// Mock oracle conmtract accessor
		var oracle *mockContractOracleAccessor = new(mockContractOracleAccessor)
		oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
		oracle.On("FinalizationPeriodSeconds").Return(faultProofWindow, nil)
		oracle.On("GetL2Output", big.NewInt(0)).Return(chain.L2Output{
			OutputRoot:    randHash().String(),
			L1Timestamp:   1000000,
			L2BlockNumber: latestL2BlockNumber,
			L2OutputIndex: 2,
		}, nil)
		oracle.On("GetL2Output", big.NewInt(1)).Return(chain.L2Output{
			OutputRoot:    randHash().String(),
			L1Timestamp:   1000000,
			L2BlockNumber: latestL2BlockNumber,
			L2OutputIndex: 2,
		}, nil)

		fd = &faultdetector.FaultDetector{
			Ctx:                    ctx,
			Logger:                 logger,
			ErrorChan:              erroChan,
			Wg:                     wg,
			Metrics:                metrics,
			L1RpcApi:               l1RpcApi,
			L2RpcApi:               l2RpcApi,
			OracleContractAccessor: oracle,
			FaultProofWindow:       faultProofWindow,
			CurrentOutputIndex:     currentOutputIndex,
			Diverged:               false,
			Ticker:                 time.NewTicker(2 * time.Second),
			QuitTickerChan:         make(chan struct{}),
			Notification:           &notification.Notification{},
		}
	}

	return fd
}

func prepareConfig(t *testing.T) *config.Config {
	serverPort, _ := strconv.Atoi(fmt.Sprintf("%d", port))

	return &config.Config{
		System: &config.System{
			LogLevel: "info",
		},
		Api: &config.Api{
			Server: &config.Server{
				Host: host,
				Port: uint(serverPort),
			},
			BasePath:         "/api",
			RegisterVersions: []string{"v1"},
		},
		FaultDetectorConfig: &config.FaultDetectorConfig{
			L1RPCEndpoint:                 l1RpcApi,
			L2RPCEndpoint:                 l2RpcApi,
			StartBatchIndex:               -1,
			L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
		},
		Notification: &config.Notification{
			Enable: true,
		},
	}
}

func parseMetricRes(input *strings.Reader) []parseMetric {
	parser := &expfmt.TextParser{}
	metricFamilies, _ := parser.TextToMetricFamilies(input)

	var parsedOutput []parseMetric
	for _, metricFamily := range metricFamilies {
		for _, m := range metricFamily.GetMetric() {
			metric := make(map[string]interface{})
			for _, label := range m.GetLabel() {
				metric[label.GetName()] = label.GetValue()
			}
			switch metricFamily.GetType() {
			case promClient.MetricType_COUNTER:
				metric["value"] = m.GetCounter().GetValue()
			case promClient.MetricType_GAUGE:
				metric["value"] = m.GetGauge().GetValue()
			}
			parsedOutput = append(parsedOutput, parseMetric{
				metricFamily.GetName(): metric,
			})
		}
	}

	return parsedOutput
}

func TestMain_E2E(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := http.DefaultClient

	tests := []struct {
		name      string
		mock      bool
		assertion func(float64, error)
	}{
		{
			name: "should start application with no faults detected",
			mock: false,
			assertion: func(isStateMismatch float64, err error) {
				const expected float64 = 0
				assert.Equal(t, isStateMismatch, expected)
				slackNotificationClient.AssertNotCalled(t, "PostMessageContext")
			},
		},
		{
			name: "should start application with faults detected",
			mock: true,
			assertion: func(isStateMismatch float64, err error) {
				const expected float64 = 1
				assert.Equal(t, isStateMismatch, expected)
				slackNotificationClient.AssertCalled(t, "PostMessageContext")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			wg := sync.WaitGroup{}
			logger, _ := log.NewDefaultProductionLogger()
			errorChan := make(chan error)
			registry := prometheus.NewRegistry()
			testConfig := prepareConfig(&testing.T{})
			testServer := prepareHTTPServer(&testing.T{}, ctx, logger, &wg, errorChan)
			testFaultDetector := prepareFaultDetector(&testing.T{}, ctx, logger, &wg, registry, testConfig, errorChan, tt.mock)
			testNotificationService := prepareNotification(&testing.T{}, ctx, logger)

			// Register handler
			testServer.Router.GET("/status", v1.GetStatus)
			testServer.RegisterHandler("GET", "/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry, ProcessStartTime: time.Now()}))

			app := &App{
				ctx:           ctx,
				logger:        logger,
				errChan:       errorChan,
				config:        testConfig,
				wg:            &wg,
				apiServer:     testServer,
				faultDetector: testFaultDetector,
				notification:  testNotificationService,
			}

			time.AfterFunc(5*time.Second, func() {
				statusEndpoint := fmt.Sprintf("http://%s:%d/status", host, port)
				req, err := http.NewRequest(http.MethodGet, statusEndpoint, nil)
				assert.NoError(t, err)
				res, err := client.Do(req)
				assert.NoError(t, err)
				assert.Equal(t, 200, res.StatusCode)

				metricsEndpoint := fmt.Sprintf("http://%s:%d/metrics", host, port)
				req, err = http.NewRequest(http.MethodGet, metricsEndpoint, nil)
				assert.NoError(t, err)
				assert.Equal(t, 200, res.StatusCode)

				res, err = client.Do(req)
				assert.NoError(t, err)
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				parsedMetric := parseMetricRes(strings.NewReader(string(body)))
				for _, m := range parsedMetric {
					if m[faultDetectorStateMismatch] != nil {
						isStateMismatch := m[faultDetectorStateMismatch]["value"].(float64)
						tt.assertion(isStateMismatch, nil)
					}
				}

				app.stop()
				wg.Done()
			})

			wg.Add(1)
			go func() {
				app.Start()
			}()
			wg.Wait()
		})
	}
}
