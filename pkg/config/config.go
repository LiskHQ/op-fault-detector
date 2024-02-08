// Package config implements the logic to read the config file (either default or user-supplied) and unmarshal it a struct for developer convenience.
package config

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/LiskHQ/op-fault-detector/pkg/utils"
	"go.uber.org/multierr"
)

var (
	providerEndpointRegex = regexp.MustCompile(`^(http|https|ws|wss)://`)
	addressRegex          = regexp.MustCompile(`(\b0x[a-f0-9]{40}\b)`)
	hostRegex             = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	basePathRegex         = regexp.MustCompile(`^/?api$`)
	registerVersionRegex  = regexp.MustCompile(`^v[1-9]\d*$`)
)

// Config struct is used to store the contents of the parsed config file.
// The properties (sub-properties) should map on-to-one with the config file.
type Config struct {
	System              *System              `mapstructure:"system"`
	Api                 *Api                 `mapstructure:"api"`
	FaultDetectorConfig *FaultDetectorConfig `mapstructure:"fault_detector"`
	Notification        *Notification        `mapstructure:"notification"`
}

// System struct is used to store the contents of the 'system' property from the parsed config file.
type System struct {
	LogLevel string `mapstructure:"log_level"`
}

// Api struct is used to store the contents of the 'api' property from the parsed config file.
type Api struct {
	Server           *Server  `mapstructure:"server"`
	BasePath         string   `mapstructure:"base_path"`
	RegisterVersions []string `mapstructure:"register_versions"`
}

// Server struct is used to store the contents of the 'api.server' sub-property from the parsed config file.
type Server struct {
	Host string `mapstructure:"host"`
	Port uint   `mapstructure:"port"`
}

// FaultDetector struct is used to store the contents of the 'fault_detector' property from the parsed config file.
type FaultDetectorConfig struct {
	L1RPCEndpoint                 string `mapstructure:"l1_rpc_endpoint"`
	L2RPCEndpoint                 string `mapstructure:"l2_rpc_endpoint"`
	Startbatchindex               int64  `mapstructure:"start_batch_index"`
	L2OutputOracleContractAddress string `mapstructure:"l2_output_oracle_contract_address"`
}

// SlackConfig struct is used to store slack configurations from the parsed config file.
type SlackConfig struct {
	ChannelID string `mapstructure:"channel_id"`
}

// notificationConfig struct is used to store notification configurations from the parsed config file.
type Notification struct {
	Slack  *SlackConfig `mapstructure:"slack"`
	Enable bool         `mapstructure:"enable"`
}

func formatError(validationErrors error) error {
	if validationErrors == nil {
		return nil
	}

	// Beautify the error message
	validationErrorSplits := strings.Split(validationErrors.Error(), ";")
	formattedErrorStr := strings.Join(validationErrorSplits, "\n\t-")

	return fmt.Errorf("fix the following %d config validation fail(s) to continue:\n\t- %s", len(validationErrorSplits), formattedErrorStr)
}

// Validate runs validations against an instance of the Config struct and returns an error when applicable.
func (c *Config) Validate() error {
	var validationErrors error

	sysConfigError := c.System.Validate()
	apiConfigError := c.Api.Validate()
	fdConfigError := c.FaultDetectorConfig.Validate()

	validationErrors = multierr.Combine(sysConfigError, apiConfigError, fdConfigError)

	return formatError(validationErrors)
}

// Validate runs validations against an instance of the System struct and returns an error when applicable.
func (c *System) Validate() error {
	var validationErrors error

	allowedLogLevels := []string{log.LevelTrace, log.LevelDebug, log.LevelInfo, log.LevelWarn, log.LevelError, log.LevelFatal}
	if !utils.Contains(allowedLogLevels, c.LogLevel) {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("system.log_level expected one of %s, received: '%s'", allowedLogLevels, c.LogLevel))
	}

	return validationErrors
}

// Validate runs validations against an instance of the Api struct and returns an error when applicable.
func (c *Api) Validate() error {
	var validationErrors error

	validationErrors = multierr.Append(validationErrors, c.Server.Validate())

	basePath := c.BasePath
	basePathMatched := basePathRegex.MatchString(basePath)
	if !basePathMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.base_path expected to match regex: `%s`, received: '%s'", basePathRegex.String(), basePath))
	}

	registerVersions := c.RegisterVersions
	for _, version := range registerVersions {
		registerVersionMatched := registerVersionRegex.MatchString(version)
		if !registerVersionMatched {
			validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.register_versions entry expected to match regex: `%s`, received: '%s'", registerVersionRegex.String(), version))
		}
	}

	return validationErrors
}

// Validate runs validations against an instance of the Server struct and returns an error when applicable.
func (c *Server) Validate() error {
	var validationErrors error

	host := c.Host
	hostMatched := hostRegex.MatchString(host)
	if !hostMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.server.host expected to match regex: `%s`, received: '%s'", hostRegex.String(), host))
	}

	port := c.Port
	minPortNum := uint(0)
	maxPortNum := uint(math.Pow(2, 16) - 1)
	if port < minPortNum || port > maxPortNum {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", minPortNum, maxPortNum, port))
	}

	return validationErrors
}

// Validate runs validations against an instance of the FaultDetector struct and returns an error when applicable.
func (c *FaultDetectorConfig) Validate() error {
	var validationErrors error

	l1ProviderMatched := providerEndpointRegex.MatchString(c.L1RPCEndpoint)
	if !l1ProviderMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("faultdetector.l1_rpc_endpoint expected to match regex: `%s`, received: '%s'", providerEndpointRegex.String(), c.L1RPCEndpoint))
	}
	l2ProviderMatched := providerEndpointRegex.MatchString(c.L2RPCEndpoint)
	if !l2ProviderMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("faultdetector.l2_rpc_endpoint expected to match regex: `%s`, received: '%s'", providerEndpointRegex.String(), c.L2RPCEndpoint))
	}

	addressMatched := addressRegex.MatchString(c.L2OutputOracleContractAddress)
	if !addressMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("faultdetector.l2_output_oracle_contract_address expected to match regex: `%s`, received: '%s'", addressRegex.String(), c.L2OutputOracleContractAddress))
	}

	return validationErrors
}
