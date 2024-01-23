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
				`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
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
					`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
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
			want: fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", `^/?api$`, "apis"),
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
			want: fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", `^v[1-9]\d*$`, "b1"),
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
			want: fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", `^v[1-9]\d*$`, "c1"),
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
				`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
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
					`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
					"127.0.0.256",
				),
				fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 99999),
				fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", `^/?api$`, "api/"),
				fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", `^v[1-9]\d*$`, "version1"),
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

// TODO: Update test table when implementing the fault detector
func TestValidate_FaultDetector(t *testing.T) {
	testCases := []struct {
		name   string
		config *FaultDetector
		want   error
	}{}

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
				fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", `^/?api$`, "api/"),
				fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", `^v[1-9]\d*$`, "version1"),
			),
			want: fmt.Errorf(
				"fix the following 3 config validation fail(s) to continue:\n\t- api.server.port expected in range: %d - %d, received: %d\n\t- api.base_path expected to match regex: `%s`, received: '%s'\n\t- api.register_versions entry expected to match regex: `%s`, received: '%s'",
				uint(0), uint(math.Pow(2, 16)-1), 23232323,
				`^/?api$`, "api/",
				`^v[1-9]\d*$`, "version1",
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
				&FaultDetector{},
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
				&FaultDetector{},
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
				&FaultDetector{},
			},
			want: formatError(
				multierr.Combine(
					fmt.Errorf(
						"api.server.host expected to match regex: `%v`, received: '%v'",
						`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`,
						"127.0.0.256",
					),
					fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", uint(0), uint(math.Pow(2, 16)-1), 99999),
					fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", `^/?api$`, "api/"),
					fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", `^v[1-9]\d*$`, "version1"),
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
