package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promClient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	host = "0.0.0.0"
	port = 8080
)

type mockChainAPIClient struct {
	mock.Mock
}

func (m *mockChainAPIClient) GetBlockHeaderByNumber(ctx context.Context) (*types.Header, error) {
	called := m.MethodCalled("GetBlockHeaderByNumber", ctx)
	return called.Get(0).(*types.Header), called.Error(1)
}

func (c *mockChainAPIClient) BlockNumber(ctx context.Context) (uint64, error) {
	ret := c.Called(ctx)

	return ret.Get(0).(uint64), ret.Error(1)
}

func (m *mockChainAPIClient) GetProof(ctx context.Context) (*chain.ProofResponse, error) {
	called := m.MethodCalled("GetProof", ctx)
	return called.Get(0).(*chain.ProofResponse), called.Error(1)
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

func prepareFaultDetector(t *testing.T, ctx context.Context, logger log.Logger, reg *prometheus.Registry, config *config.Config, wg *sync.WaitGroup, erroChan chan error, mock bool) *faultdetector.FaultDetector {
	var fd *faultdetector.FaultDetector
	if !mock {
		fd, _ = faultdetector.NewFaultDetector(ctx, logger, erroChan, wg, config.FaultDetectorConfig, reg, &notification.Notification{})
	} else {
		// TODO: Mock chains APIs
		l1RpcApi, _ := chain.GetAPIClient(ctx, "https://rpc.notadegen.com/eth", logger)
		l2RpcApi := new(mockChainAPIClient)

		fd = &faultdetector.FaultDetector{
			Ctx:                    ctx,
			Logger:                 logger,
			ErrorChan:              erroChan,
			Wg:                     wg,
			Metrics:                &faultdetector.FaultDetectorMetrics{},
			L1RpcApi:               l1RpcApi,
			L2RpcApi:               l2RpcApi,
			OracleContractAccessor: &chain.OracleAccessor{},
			FaultProofWindow:       60480,
			CurrentOutputIndex:     1,
			Diverged:               false,
			Ticker:                 time.NewTicker(2 * time.Second),
			QuitTickerChan:         make(chan struct{}),
			Notification:           &notification.Notification{},
		}
	}

	return fd
}

func prepareConfig(t *testing.T) *config.Config {
	return &config.Config{
		System: &config.System{
			LogLevel: "info",
		},
		Api: &config.Api{
			Server: &config.Server{
				Host: "0.0.0.0",
				Port: 8080,
			},
			BasePath:         "/api",
			RegisterVersions: []string{"v1"},
		},
		FaultDetectorConfig: &config.FaultDetectorConfig{
			L1RPCEndpoint:                 "https://rpc.notadegen.com/eth",
			L2RPCEndpoint:                 "https://mainnet.optimism.io/",
			Startbatchindex:               -1,
			L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
		},
		Notification: &config.Notification{
			Enable: false,
		},
	}
}

func parseMetricRes(input *strings.Reader) []map[string]map[string]interface{} {
	parser := &expfmt.TextParser{}
	metricFamilies, _ := parser.TextToMetricFamilies(input)

	var parsedOutput []map[string]map[string]interface{}
	for _, value := range metricFamilies {
		for _, m := range value.GetMetric() {
			metric := make(map[string]interface{})
			for _, label := range m.GetLabel() {
				metric[label.GetName()] = label.GetValue()
			}
			switch value.GetType() {
			case promClient.MetricType_COUNTER:
				metric["value"] = m.GetCounter().GetValue()
			case promClient.MetricType_GAUGE:
				metric["value"] = m.GetGauge().GetValue()
			}
			parsedOutput = append(parsedOutput, map[string]map[string]interface{}{
				value.GetName(): metric,
			})
		}
	}

	return parsedOutput
}

func TestMain_E2E(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	wg := sync.WaitGroup{}
	logger, _ := log.NewDefaultProductionLogger()
	client := http.DefaultClient
	erroChan := make(chan error)
	registry := prometheus.NewRegistry()
	testConfig := prepareConfig(&testing.T{})
	testServer := prepareHTTPServer(&testing.T{}, ctx, logger, &wg, erroChan)
	testFaultDetector := prepareFaultDetector(&testing.T{}, ctx, logger, registry, testConfig, &wg, erroChan, false)
	testFaultDetectorMocked := prepareFaultDetector(&testing.T{}, ctx, logger, registry, testConfig, &wg, erroChan, true)

	// Register handler
	testServer.Router.GET("/status", v1.GetStatus)
	testServer.RegisterHandler("GET", "/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry, ProcessStartTime: time.Now()}))

	tests := []struct {
		name      string
		App       App
		assertion func(float64, error)
	}{
		{
			name: "should start application with no faults detected",
			App: App{
				ctx:           ctx,
				logger:        logger,
				errChan:       erroChan,
				config:        testConfig,
				apiServer:     testServer,
				faultDetector: testFaultDetector,
				notification:  &notification.Notification{},
				wg:            &wg,
			},
			assertion: func(isStateMismatch float64, err error) {
				var expected float64 = 0
				assert.Equal(t, isStateMismatch, expected)
			},
		},
		{
			name: "should start application with faults detected",
			App: App{
				ctx:           ctx,
				logger:        logger,
				errChan:       erroChan,
				config:        testConfig,
				apiServer:     testServer,
				faultDetector: testFaultDetectorMocked,
				notification:  &notification.Notification{},
				wg:            &wg,
			},
			assertion: func(isStateMismatch float64, err error) {
				var expected float64 = 1
				assert.Equal(t, isStateMismatch, expected)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				ctx:           tt.App.ctx,
				logger:        tt.App.logger,
				errChan:       tt.App.errChan,
				config:        tt.App.config,
				wg:            tt.App.wg,
				apiServer:     tt.App.apiServer,
				faultDetector: tt.App.faultDetector,
				notification:  tt.App.notification,
			}

			time.AfterFunc(5*time.Second, func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/status", host, port), nil)
				assert.NoError(t, err)
				res, err := client.Do(req)
				assert.NoError(t, err)
				assert.Equal(t, 200, res.StatusCode)

				req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/metrics", host, port), nil)
				assert.NoError(t, err)
				assert.Equal(t, 200, res.StatusCode)

				res, err = client.Do(req)
				assert.NoError(t, err)
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				parsedMetric := parseMetricRes(strings.NewReader(string(body)))
				for _, m := range parsedMetric {
					if m["fault_detector_is_state_mismatch"] != nil {
						isStateMismatch := m["fault_detector_is_state_mismatch"]["value"].(float64)
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