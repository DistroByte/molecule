package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigurationError(t *testing.T) {
	t.Run("error without cause", func(t *testing.T) {
		err := NewConfigurationError("test message", nil)
		assert.Equal(t, "configuration error: test message", err.Error())
		assert.Nil(t, err.Unwrap())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewConfigurationError("test message", cause)
		assert.Equal(t, "configuration error: test message - underlying error", err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestServiceError(t *testing.T) {
	t.Run("error without cause", func(t *testing.T) {
		err := NewServiceError("test-service", "test message", nil)
		assert.Equal(t, "service error [test-service]: test message", err.Error())
		assert.Nil(t, err.Unwrap())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewServiceError("test-service", "test message", cause)
		assert.Equal(t, "service error [test-service]: test message - underlying error", err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestDomainErrors(t *testing.T) {
	testCases := []struct {
		name  string
		err   error
		wants string
	}{
		{"ErrConfigNotFound", ErrConfigNotFound, "configuration not found"},
		{"ErrInvalidConfig", ErrInvalidConfig, "invalid configuration"},
		{"ErrNomadClientFailed", ErrNomadClientFailed, "failed to create nomad client"},
		{"ErrServiceNotFound", ErrServiceNotFound, "service not found"},
		{"ErrAllocationFailed", ErrAllocationFailed, "allocation operation failed"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized access"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wants, tc.err.Error())
		})
	}
}