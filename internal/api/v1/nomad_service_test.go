package v1

import (
	"errors"
	"testing"

	generated "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
)

// NomadClientInterface defines the interface needed for testing
type NomadClientInterface interface {
	Allocations() AllocationsInterface
	Jobs() JobsInterface
	Nodes() NodesInterface
}

type AllocationsInterface interface {
	List(q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error)
	Info(allocID string, q *api.QueryOptions) (*api.Allocation, *api.QueryMeta, error)
	Restart(alloc *api.Allocation, taskName string, q *api.WriteOptions) error
}

type JobsInterface interface {
	Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error)
}

type NodesInterface interface {
	Info(nodeID string, q *api.QueryOptions) (*api.Node, *api.QueryMeta, error)
}

// Mock implementations
type MockNomadClient struct {
	shouldFail  bool
	allocations []*api.AllocationListStub
}

func (m *MockNomadClient) Allocations() AllocationsInterface {
	return &MockAllocations{shouldFail: m.shouldFail, allocations: m.allocations}
}

func (m *MockNomadClient) Jobs() JobsInterface {
	return &MockJobs{shouldFail: m.shouldFail}
}

func (m *MockNomadClient) Nodes() NodesInterface {
	return &MockNodes{shouldFail: m.shouldFail}
}

type MockAllocations struct {
	shouldFail  bool
	allocations []*api.AllocationListStub
}

func (m *MockAllocations) List(q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error) {
	if m.shouldFail {
		return nil, nil, errors.New("mock error")
	}
	return m.allocations, &api.QueryMeta{}, nil
}

func (m *MockAllocations) Info(allocID string, q *api.QueryOptions) (*api.Allocation, *api.QueryMeta, error) {
	if m.shouldFail {
		return nil, nil, errors.New("mock error")
	}
	return &api.Allocation{}, &api.QueryMeta{}, nil
}

func (m *MockAllocations) Restart(alloc *api.Allocation, taskName string, q *api.WriteOptions) error {
	if m.shouldFail {
		return errors.New("mock error")
	}
	return nil
}

type MockJobs struct {
	shouldFail bool
}

func (m *MockJobs) Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error) {
	if m.shouldFail {
		return nil, nil, errors.New("mock error")
	}
	return &api.Job{}, &api.QueryMeta{}, nil
}

type MockNodes struct {
	shouldFail bool
}

func (m *MockNodes) Info(nodeID string, q *api.QueryOptions) (*api.Node, *api.QueryMeta, error) {
	if m.shouldFail {
		return nil, nil, errors.New("mock error")
	}
	return &api.Node{}, &api.QueryMeta{}, nil
}

// Test functions with simpler interface
func TestMakeUnique(t *testing.T) {
	input := []generated.ServiceUrl{
		{Service: "service1", Url: "https://service1.com"},
		{Service: "service2", Url: "https://service2.com"},
		{Service: "service1", Url: "https://service1-duplicate.com"}, // Should be deduplicated
		{Service: "service3", Url: "https://service3.com"},
	}

	result := makeUnique(input)

	assert.Len(t, result, 3)
	
	// Check that services are sorted alphabetically
	assert.Equal(t, "service1", result[0].Service)
	assert.Equal(t, "service2", result[1].Service)
	assert.Equal(t, "service3", result[2].Service)

	// Check that first occurrence is kept
	assert.Equal(t, "https://service1.com", result[0].Url)
}

func TestRegexes(t *testing.T) {
	t.Run("traefik rule regex", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"traefik.http.routers.test.rule", true},
			{"traefik.http.routers.web-service.rule", true},
			{"traefik.http.routers.123.rule", true},
			{"traefik.enable=true", false},
			{"traefik.port=80", false},
			{"random-tag", false},
		}

		for _, tc := range testCases {
			result := traefikRuleTagRegex.MatchString(tc.input)
			assert.Equal(t, tc.expected, result, "input: %s", tc.input)
		}
	})

	t.Run("icon tag regex", func(t *testing.T) {
		testCases := []struct {
			input     string
			expected  bool
			iconValue string
		}{
			{"icon=test-icon", true, "test-icon"},
			{"icon=mdi:home", true, "mdi:home"},
			{"icon=", true, ""},
			{"not-icon=test", true, "test"},   // This matches because the regex looks for "icon=" anywhere
			{"prefix-icon=test", true, "test"}, // This matches because the regex looks for "icon=" anywhere
			{"some-tag=value", false, ""},     // This doesn't contain "icon="
		}

		for _, tc := range testCases {
			matches := iconTagRegex.FindStringSubmatch(tc.input)
			if tc.expected {
				assert.Len(t, matches, 2, "input: %s", tc.input)
				assert.Equal(t, tc.iconValue, matches[1], "input: %s", tc.input)
			} else {
				assert.Nil(t, matches, "input: %s", tc.input)
			}
		}
	})
}

func TestNomadService_NewNomadService(t *testing.T) {
	// Test the constructor
	mockClient := &api.Client{}
	standardURLs := []generated.ServiceUrl{
		{Service: "test", Url: "https://test.com"},
	}

	service := NewNomadService(mockClient, standardURLs)
	assert.NotNil(t, service)
	assert.IsType(t, &NomadService{}, service)
}

// Test basic functionality that doesn't require complex mocking
func TestNomadService_CreationAndInterfaces(t *testing.T) {
	// Test the constructor with actual client (but don't call methods that need API)
	mockClient := &api.Client{}
	standardURLs := []generated.ServiceUrl{
		{Service: "test", Url: "https://test.com"},
	}

	service := NewNomadService(mockClient, standardURLs)
	
	assert.NotNil(t, service)
	assert.IsType(t, &NomadService{}, service)
	
	// Test that the service implements the interface
	var _ NomadServiceInterface = service
}