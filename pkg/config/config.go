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

// Config struct is used to store the contents of the parsed config file.
// The properties (sub-properties) should map on-to-one with the config file.
type Config struct {
	System        *System        `mapstructure:"system"`
	Api           *Api           `mapstructure:"api"`
	FaultDetector *FaultDetector `mapstructure:"fault_detector"`
}

type System struct {
	LogLevel string `mapstructure:"log_level"`
}

type Api struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port uint   `mapstructure:"port"`
	} `mapstructure:"server"`
	BasePath         string   `mapstructure:"base_path"`
	RegisterVersions []string `mapstructure:"register_versions"`
}

type FaultDetector struct {
}

func (c *Config) Validate() error {
	var validationErrors error

	sysConfigError := validateSystemConfig(c)
	apiConfigError := validateApiConfig(c)
	fdConfigError := validateFaultDetectorConfig(c)

	validationErrors = multierr.Combine(sysConfigError, apiConfigError, fdConfigError)

	if validationErrors != nil {
		// Beautify the error message
		validationErrorSplits := strings.Split(validationErrors.Error(), ";")
		formattedErrorStr := strings.Join(validationErrorSplits, "\n\t- ")

		return fmt.Errorf("fix the following %d config validation fail(s) to continue:\n\t - %s", len(validationErrorSplits), formattedErrorStr)
	}

	return validationErrors
}

func validateSystemConfig(c *Config) error {
	var validationErrors error

	allowedLogLevels := []string{log.LevelTrace, log.LevelDebug, log.LevelInfo, log.LevelWarn, log.LevelError, log.LevelFatal}
	if !utils.Contains(allowedLogLevels, c.System.LogLevel) {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("system.log_level expected one of %s, received: '%s'", allowedLogLevels, c.System.LogLevel))
	}

	return validationErrors
}

func validateApiConfig(c *Config) error {
	var validationErrors error

	validationErrors = multierr.Append(validationErrors, validateApiServerConfig(c))

	basePath := c.Api.BasePath
	basePathRegex := `^/?api$`
	basePathMatched, _ := regexp.MatchString(basePathRegex, basePath)
	if !basePathMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.base_path expected to match regex: `%v`, received: '%s'", basePathRegex, basePath))
	}

	registerVersions := c.Api.RegisterVersions
	registerVersionRegex := `^v[1-9]\d*$`
	registerVersionRegexCompiled, _ := regexp.Compile(registerVersionRegex)
	for _, version := range registerVersions {
		registerVersionMatched := registerVersionRegexCompiled.MatchString(version)
		if !registerVersionMatched {
			validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.register_versions entry expected to match regex: `%v`, received: '%s'", registerVersionRegex, version))
		}
	}

	return validationErrors
}

func validateApiServerConfig(c *Config) error {
	var validationErrors error

	host := c.Api.Server.Host
	hostRegex := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	hostMatched, _ := regexp.MatchString(hostRegex, host)
	if !hostMatched {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.server.host expected to match regex: `%v`, received: '%s'", hostRegex, host))
	}

	port := c.Api.Server.Port
	minPortNum := uint(0)
	maxPortNum := uint(math.Pow(2, 16) - 1)
	if port < minPortNum || port > maxPortNum {
		validationErrors = multierr.Append(validationErrors, fmt.Errorf("api.server.port expected in range: %d - %d, received: %d", minPortNum, maxPortNum, port))
	}

	return validationErrors
}

func validateFaultDetectorConfig(c *Config) error {
	return nil
}
