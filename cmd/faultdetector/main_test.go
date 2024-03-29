package main

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/api"
	"github.com/LiskHQ/op-fault-detector/pkg/chain"
	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/faultdetector"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/LiskHQ/op-fault-detector/pkg/utils/notification"
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
	host                                = "127.0.0.1"
	port                                = 8088
	faultProofWindow                    = 1000
	currentOutputIndex                  = 1
	l1RpcApi                            = "https://rpc.notadegen.com/eth"
	l2RpcApi                            = "https://mainnet.optimism.io/"
	faultDetectorStateMismatchMetricKey = "fault_detector_is_state_mismatch"
	metricValue                         = "value"
	postMessageContextFnName            = "PostMessageContext"
	channelID                           = "TestChannelID"
)

type parsedMetricMap map[string]map[string]interface{}

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
	called := o.MethodCalled(postMessageContextFnName)
	return called.Get(0).(string), called.Get(1).(string), called.Error(2)
}

var slackNotificationClient *mockSlackClient = new(mockSlackClient)

func randHash() (out common.Hash) {
	_, _ = crand.Read(out[:])
	return out
}

func randTimestamp() (out uint64) {
	timestamp := uint64(rand.Int63n(time.Now().Unix()))
	return timestamp
}

func prepareHTTPServer(t *testing.T, ctx context.Context, logger log.Logger, config *config.Config, wg *sync.WaitGroup, erroChan chan error) *api.HTTPServer {
	testServer := api.NewHTTPServer(ctx, logger, wg, config, erroChan)
	return testServer
}

func prepareNotification(t *testing.T, ctx context.Context, logger log.Logger, config *config.Config) *notification.Notification {
	slackNotificationClient.On(postMessageContextFnName).Return(channelID, "1234569.1000", nil)
	testNotificationService := notification.GetNotification(ctx, logger, slackNotificationClient, config.Notification)
	return testNotificationService
}

func prepareFaultDetector(t *testing.T, ctx context.Context, logger log.Logger, testNotificationService *notification.Notification, wg *sync.WaitGroup, reg *prometheus.Registry, config *config.Config, erroChan chan error, mock bool) *faultdetector.FaultDetector {
	var fd *faultdetector.FaultDetector
	if !mock {
		fd, _ = faultdetector.NewFaultDetector(ctx, logger, erroChan, wg, config.FaultDetectorConfig, reg, testNotificationService)
	} else {
		mx := new(sync.RWMutex)
		metrics := faultdetector.NewFaultDetectorMetrics(reg)

		// Create chain API clients
		l1RpcApi, err := chain.GetAPIClient(ctx, l1RpcApi, logger)
		if err != nil {
			panic(err)
		}
		l2RpcApi, err := chain.GetAPIClient(ctx, l2RpcApi, logger)
		if err != nil {
			panic(err)
		}

		latestL2BlockNumber, err := l2RpcApi.GetLatestBlockNumber(ctx)
		if err != nil {
			panic(err)
		}

		// Mock oracle contract accessor
		var oracle *mockContractOracleAccessor = new(mockContractOracleAccessor)
		oracle.On("GetNextOutputIndex").Return(big.NewInt(2), nil)
		oracle.On("FinalizationPeriodSeconds").Return(faultProofWindow, nil)
		oracle.On("GetL2Output", big.NewInt(0)).Return(chain.L2Output{
			OutputRoot:    randHash().String(),
			L1Timestamp:   randTimestamp(),
			L2BlockNumber: latestL2BlockNumber,
			L2OutputIndex: 2,
		}, nil)
		oracle.On("GetL2Output", big.NewInt(1)).Return(chain.L2Output{
			OutputRoot:    randHash().String(),
			L1Timestamp:   randTimestamp(),
			L2BlockNumber: latestL2BlockNumber,
			L2OutputIndex: 2,
		}, nil)

		fd = faultdetector.GetFaultDetector(ctx, logger, l1RpcApi, l2RpcApi, oracle, faultProofWindow, currentOutputIndex, metrics, testNotificationService, false, wg, erroChan, mx)
	}

	return fd
}

func prepareConfig(t *testing.T) *config.Config {
	serverPort, err := strconv.Atoi(fmt.Sprintf("%d", port))
	if err != nil {
		panic(err)
	}

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
			Slack: &config.SlackConfig{
				ChannelID: channelID,
			},
		},
	}
}

func parseMetricRes(input *strings.Reader) []parsedMetricMap {
	parser := &expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(input)
	if err != nil {
		panic(err)
	}

	var parsedOutput []parsedMetricMap
	for _, metricFamily := range metricFamilies {
		for _, m := range metricFamily.GetMetric() {
			metric := make(map[string]interface{})
			for _, label := range m.GetLabel() {
				metric[label.GetName()] = label.GetValue()
			}
			switch metricFamily.GetType() {
			case promClient.MetricType_COUNTER:
				metric[metricValue] = m.GetCounter().GetValue()
			case promClient.MetricType_GAUGE:
				metric[metricValue] = m.GetGauge().GetValue()
			}
			parsedOutput = append(parsedOutput, parsedMetricMap{
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
				slackNotificationClient.AssertNotCalled(t, postMessageContextFnName)
			},
		},
		{
			name: "should start application with faults detected",
			mock: true,
			assertion: func(isStateMismatch float64, err error) {
				const expected float64 = 1
				assert.Equal(t, isStateMismatch, expected)
				slackNotificationClient.AssertCalled(t, postMessageContextFnName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			wg := sync.WaitGroup{}
			logger, err := log.NewDefaultProductionLogger()
			if err != nil {
				panic(err)
			}

			errorChan := make(chan error)
			registry := prometheus.NewRegistry()
			testConfig := prepareConfig(&testing.T{})
			testServer := prepareHTTPServer(&testing.T{}, ctx, logger, testConfig, &wg, errorChan)
			testNotificationService := prepareNotification(&testing.T{}, ctx, logger, testConfig)
			testFaultDetector := prepareFaultDetector(&testing.T{}, ctx, logger, testNotificationService, &wg, registry, testConfig, errorChan, tt.mock)

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
				statusEndpoint := fmt.Sprintf("http://%s:%d/api/v1/status", host, port)
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
					if m[faultDetectorStateMismatchMetricKey] != nil {
						isStateMismatch := m[faultDetectorStateMismatchMetricKey][metricValue].(float64)
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
