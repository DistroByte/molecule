package domain

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	ErrConfigNotFound    = errors.New("configuration not found")
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrNomadClientFailed = errors.New("failed to create nomad client")
	ErrServiceNotFound   = errors.New("service not found")
	ErrAllocationFailed  = errors.New("allocation operation failed")
	ErrUnauthorized      = errors.New("unauthorized access")
)

// ConfigurationError represents configuration-related errors
type ConfigurationError struct {
	Message string
	Cause   error
}

func (e *ConfigurationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("configuration error: %s - %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("configuration error: %s", e.Message)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Cause
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string, cause error) *ConfigurationError {
	return &ConfigurationError{
		Message: message,
		Cause:   cause,
	}
}

// ServiceError represents service-related errors
type ServiceError struct {
	Service string
	Message string
	Cause   error
}

func (e *ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("service error [%s]: %s - %v", e.Service, e.Message, e.Cause)
	}
	return fmt.Sprintf("service error [%s]: %s", e.Service, e.Message)
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError creates a new service error
func NewServiceError(service, message string, cause error) *ServiceError {
	return &ServiceError{
		Service: service,
		Message: message,
		Cause:   cause,
	}
}
