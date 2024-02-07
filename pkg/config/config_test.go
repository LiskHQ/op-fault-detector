package config

import (
	"fmt"
	"math"
	"testing"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/magiconair/properties/assert"
	"go.uber.org/multierr"
)

func TestValidate_System(t *testing.T) {
	testCases := []struct {
		name   string
		config *System
		want   error
	}{
		{
			name: "should return nil when correct system config specified",
			config: &System{
				"info",
			},
			want: nil,
		},
		{
			name: "should return error when incorrect log level specified in system config",
			config: &System{
				"tracer",
			},
			want: fmt.Errorf(
				"system.log_level expected one of %s, received: '%s'",
				[]string{log.LevelTrace, log.LevelDebug, log.LevelInfo, log.LevelWarn, log.LevelError, log.LevelFatal},
				"tracer",
			),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Validate()
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestValidate_Server(t *testing.T) {
	testCases := []struct {
		name   string
		config *Server
		want   error
	}{
		{
			name: "should return nil when correct server config specified",
			config: &Server{
				"0.0.0.0",
				8080,
			},
			want: nil,
		},
		{
			name: "should return error when incorrect host specified in system config",
			config: &Server{
				"127.0.0.256",
				8080,
			},
			want: fmt.Errorf(
				"api.server.host expected to match regex: `%v`, received: '%v'",
				hostRegex.String(),
				"127.0.0.256",
			),
		},
		{
			name: "should return error when incorrect port specified in system config",
			config: &Server{
				"127.0.0.1",
				99999,
			},
			want: fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 99999),
		},
		{
			name: "should return error when incorrect host and port specified in system config",
			config: &Server{
				"127.0.0.256",
				99999,
			},
			want: multierr.Append(
				fmt.Errorf(
					"api.server.host expected to match regex: `%v`, received: '%v'",
					hostRegex.String(),
					"127.0.0.256",
				),
				fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 99999),
			),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Validate()
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestValidate_Api(t *testing.T) {
	testCases := []struct {
		name   string
		config *Api
		want   error
	}{
		{
			name: "should return nil when correct api config specified",
			config: &Api{
				&Server{
					"0.0.0.0",
					8080,
				},
				"/api",
				[]string{"v1"},
			},
			want: nil,
		},
		{
			name: "should return nil when correct base_path specified without leading slash in api config",
			config: &Api{
				&Server{
					"0.0.0.0",
					8080,
				},
				"api",
				[]string{"v1"},
			},
			want: nil,
		},
		{
			name: "should return error when incorrect base_path specified in api config",
			config: &Api{
				&Server{
					"0.0.0.0",
					8080,
				},
				"apis",
				[]string{"v1"},
			},
			want: fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", basePathRegex.String(), "apis"),
		},
		{
			name: "should return error when incorrect register_versions specified in api config",
			config: &Api{
				&Server{
					"0.0.0.0",
					8080,
				},
				"/api",
				[]string{"b1"},
			},
			want: fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", registerVersionRegex.String(), "b1"),
		},
		{
			name: "should return error when one of register_versions specified is incorrect in api config",
			config: &Api{
				&Server{
					"0.0.0.0",
					8080,
				},
				"api",
				[]string{"v1", "v2", "c1"},
			},
			want: fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", registerVersionRegex.String(), "c1"),
		},
		{
			name: "should return error when server host specified is incorrect in api config",
			config: &Api{
				&Server{
					"256.255.0.8",
					8080,
				},
				"api",
				[]string{"v1"},
			},
			want: fmt.Errorf(
				"api.server.host expected to match regex: `%v`, received: '%v'",
				hostRegex.String(),
				"256.255.0.8",
			),
		},
		{
			name: "should return error when server port specified is incorrect in api config",
			config: &Api{
				&Server{
					"0.0.0.0",
					556677,
				},
				"api",
				[]string{"v1"},
			},
			want: fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 556677),
		},
		{
			name: "should return error when multiple incorrect parameters specified in system config",
			config: &Api{
				&Server{
					"127.0.0.256",
					99999,
				},
				"api/",
				[]string{"v1", "version1"},
			},
			want: multierr.Combine(
				fmt.Errorf(
					"api.server.host expected to match regex: `%v`, received: '%v'",
					hostRegex.String(),
					"127.0.0.256",
				),
				fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 99999),
				fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", basePathRegex.String(), "api/"),
				fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", registerVersionRegex.String(), "version1"),
			),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Validate()
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestValidate_FaultDetector(t *testing.T) {
	testCases := []struct {
		name   string
		config *FaultDetectorConfig
		want   error
	}{
		{
			name: "should return nil when correct http/https endpoints are given",
			config: &FaultDetectorConfig{
				L1RPCEndpoint:                 "https://xyz.com",
				L2RPCEndpoint:                 "http://xyz.com",
				Startbatchindex:               100,
				L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
			},
			want: nil,
		},
		{
			name: "should return nil when correct ws/wss endpoints are given",
			config: &FaultDetectorConfig{
				L1RPCEndpoint:                 "wss://xyz.com",
				L2RPCEndpoint:                 "ws://xyz.com",
				Startbatchindex:               100,
				L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
			},
			want: nil,
		},
		{
			name: "should return error when invalid l1 provider endpoint is given",
			config: &FaultDetectorConfig{
				L1RPCEndpoint:                 "://xyz.com",
				L2RPCEndpoint:                 "http://xyz.com",
				Startbatchindex:               100,
				L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
			},
			want: fmt.Errorf("faultdetector.l1_rpc_endpoint expected to match regex: `%s`, received: '://xyz.com'", providerEndpointRegex.String()),
		},
		{
			name: "should return error when invalid l2 provider endpoint is given",
			config: &FaultDetectorConfig{
				L1RPCEndpoint:                 "http://xyz.com",
				L2RPCEndpoint:                 "ht://xyz.com",
				Startbatchindex:               100,
				L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
			},
			want: fmt.Errorf("faultdetector.l2_rpc_endpoint expected to match regex: `%s`, received: 'ht://xyz.com'", providerEndpointRegex.String()),
		},
		{
			name: "should return error when invalid address length",
			config: &FaultDetectorConfig{
				L1RPCEndpoint:                 "http://xyz.com",
				L2RPCEndpoint:                 "http://xyz.com",
				Startbatchindex:               100,
				L2OutputOracleContractAddress: "randomAddress",
			},
			want: fmt.Errorf("faultdetector.l2_output_oracle_contract_address expected to match regex: `%s`, received: 'randomAddress'", addressRegex.String()),
		},
		{
			name: "should return error when invalid address beginning",
			config: &FaultDetectorConfig{
				L1RPCEndpoint:                 "http://xyz.com",
				L2RPCEndpoint:                 "http://xyz.com",
				Startbatchindex:               100,
				L2OutputOracleContractAddress: "xx0000000000000000000000000000000000000000",
			},
			want: fmt.Errorf("faultdetector.l2_output_oracle_contract_address expected to match regex: `%s`, received: 'xx0000000000000000000000000000000000000000'", addressRegex.String()),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Validate()
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestFormatError(t *testing.T) {
	testCases := []struct {
		name             string
		validationErrors error
		want             error
	}{
		{
			name:             "should return nil when nil error is supplied",
			validationErrors: nil,
			want:             nil,
		},
		{
			name:             "should return properly formatted error when a single error is supplied",
			validationErrors: fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 543210),
			want: fmt.Errorf(
				"fix the following 1 config validation fail(s) to continue:\n\t- api.server.port expected in range: %d - %d, received: %d",
				uint(0),
				uint(math.Pow(2, 16)-1),
				543210,
			),
		},

		{
			name: "should return properly formatted error when multiple combined errors are supplied",
			validationErrors: multierr.Combine(
				fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 23232323),
				fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", basePathRegex.String(), "api/"),
				fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", registerVersionRegex.String(), "version1"),
			),
			want: fmt.Errorf(
				"fix the following 3 config validation fail(s) to continue:\n\t- api.server.port expected in range: %d - %d, received: %d\n\t- api.base_path expected to match regex: `%s`, received: '%s'\n\t- api.register_versions entry expected to match regex: `%s`, received: '%s'",
				uint(0), uint(math.Pow(2, 16)-1), 23232323,
				basePathRegex.String(), "api/",
				registerVersionRegex.String(), "version1",
			),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatError(tc.validationErrors)
			assert.Equal(t, got, tc.want)
		})
	}
}

// TODO: Adjust the test table below after implementing the fault detector
func TestValidate_Config(t *testing.T) {
	testCases := []struct {
		name   string
		config *Config
		want   error
	}{
		{
			name: "should return nil when correct config specified",
			config: &Config{
				&System{
					"info",
				},
				&Api{
					&Server{
						"0.0.0.0",
						8080,
					},
					"/api",
					[]string{"v1"},
				},
				&FaultDetectorConfig{
					L1RPCEndpoint:                 "https://xyz.com",
					L2RPCEndpoint:                 "http://xyz.com",
					Startbatchindex:               100,
					L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
				},
				&SlackConfig{
					ChannelID: "testID",
				},
			},
			want: nil,
		},
		{
			name: "should return error when incorrect port number specified in config",
			config: &Config{
				&System{
					"info",
				},
				&Api{
					&Server{
						"0.0.0.0",
						543210,
					},
					"/api",
					[]string{"v1"},
				},
				&FaultDetectorConfig{
					L1RPCEndpoint:                 "https://xyz.com",
					L2RPCEndpoint:                 "http://xyz.com",
					Startbatchindex:               100,
					L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
				},
				&SlackConfig{
					ChannelID: "testID",
				},
			},
			want: formatError(
				fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 543210),
			),
		},
		{
			name: "should return error when incorrect port number specified in config",
			config: &Config{
				&System{
					"info",
				},
				&Api{
					&Server{
						"127.0.0.256",
						99999,
					},
					"api/",
					[]string{"v1", "version1"},
				},
				&FaultDetectorConfig{
					L1RPCEndpoint:                 "https://xyz.com",
					L2RPCEndpoint:                 "http://xyz.com",
					Startbatchindex:               100,
					L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
				},
				&SlackConfig{
					ChannelID: "testID",
				},
			},
			want: formatError(
				multierr.Combine(
					fmt.Errorf(
						"api.server.host expected to match regex: `%v`, received: '%v'",
						hostRegex.String(),
						"127.0.0.256",
					),
					fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 99999),
					fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", basePathRegex.String(), "api/"),
					fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", registerVersionRegex.String(), "version1"),
				),
			),
		},
		{
			name: "should return error when incorrect L1RPCEndpoint is provided",
			config: &Config{
				&System{
					"info",
				},
				&Api{
					&Server{
						"127.0.0.256",
						8080,
					},
					"/api",
					[]string{"v1"},
				},
				&FaultDetectorConfig{
					L1RPCEndpoint:                 "htps://xyz.com",
					L2RPCEndpoint:                 "http://xyz.com",
					Startbatchindex:               100,
					L2OutputOracleContractAddress: "0x0000000000000000000000000000000000000000",
				},
				&SlackConfig{
					ChannelID: "testID",
				},
			},
			want: formatError(
				multierr.Combine(
					fmt.Errorf(
						"api.server.host expected to match regex: `%v`, received: '%v'",
						hostRegex.String(),
						"127.0.0.256",
					),
					fmt.Errorf("faultdetector.l1_rpc_endpoint expected to match regex: `%s`, received: 'htps://xyz.com'", providerEndpointRegex.String()),
				),
			),
		},
		{
			name: "should return error when incorrect L2OutputOracleContractAddress is provided",
			config: &Config{
				&System{
					"info",
				},
				&Api{
					&Server{
						"127.0.0.256",
						8080,
					},
					"/api",
					[]string{"v1"},
				},
				&FaultDetectorConfig{
					L1RPCEndpoint:                 "https://xyz.com",
					L2RPCEndpoint:                 "htps://xyz.com",
					Startbatchindex:               100,
					L2OutputOracleContractAddress: "xx0000000000000000000000000000000000000000",
				},
				&SlackConfig{
					ChannelID: "testChannelID",
				},
			},
			want: formatError(
				multierr.Combine(
					fmt.Errorf(
						"api.server.host expected to match regex: `%v`, received: '%v'",
						hostRegex.String(),
						"127.0.0.256",
					),
					fmt.Errorf("faultdetector.l2_rpc_endpoint expected to match regex: `%s`, received: 'htps://xyz.com'", providerEndpointRegex.String()),
					fmt.Errorf("faultdetector.l2_output_oracle_contract_address expected to match regex: `%s`, received: 'xx0000000000000000000000000000000000000000'", addressRegex.String()),
				),
			),
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Validate()
			assert.Equal(t, got, tc.want)
		})
	}
}
