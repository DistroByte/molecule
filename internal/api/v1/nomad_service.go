package v1

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	generated "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/DistroByte/molecule/logger"
	"github.com/hashicorp/nomad/api"
	"github.com/jedib0t/go-pretty/v6/table"
)

// NomadService handles Nomad cluster interactions
type NomadService struct {
	nomadClient   *api.Client
	standardURLs  []generated.ServiceUrl
	mu            sync.RWMutex
}

// NomadServiceInterface defines the interface for Nomad service operations
type NomadServiceInterface interface {
	ExtractAll(print bool) ([]generated.ServiceUrl, error)
	ExtractURLs() ([]generated.ServiceUrl, error)
	ExtractHostPorts() ([]generated.ServiceUrl, error)
	ExtractServicePorts() ([]generated.ServiceUrl, error)
	GetServiceStatus(service string) (map[string]string, error)
	RestartServiceAllocations(service string) error
}

var (
	traefikRuleTagRegex = regexp.MustCompile("traefik.http.routers.*.rule")
	iconTagRegex        = regexp.MustCompile("icon=(.*)")
)

// allocationData holds the extracted data from allocations
type allocationData struct {
	serviceUrls       []generated.ServiceUrl
	hostReservedPorts []generated.ServiceUrl
	servicePorts      []generated.ServiceUrl
}

// NewNomadService creates a new NomadService instance
func NewNomadService(nomadClient *api.Client, staticUrls []generated.ServiceUrl) NomadServiceInterface {
	return &NomadService{
		nomadClient:  nomadClient,
		standardURLs: staticUrls,
	}
}

// processAllocationsData handles the common pattern of getting and processing allocations
func (s *NomadService) processAllocationsData() (*allocationData, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	data := &allocationData{
		serviceUrls:       []generated.ServiceUrl{},
		hostReservedPorts: []generated.ServiceUrl{},
		servicePorts:      []generated.ServiceUrl{},
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation, data)
	}

	return data, nil
}

// ExtractAll extracts all URLs from Nomad allocations
func (s *NomadService) ExtractAll(print bool) ([]generated.ServiceUrl, error) {
	data, err := s.processAllocationsData()
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	standardURLs := s.standardURLs
	s.mu.RUnlock()

	allUrls := []generated.ServiceUrl{}
	
	// Add service URLs
	for _, v := range data.serviceUrls {
		allUrls = append(allUrls, generated.ServiceUrl{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
			Icon:    v.Icon,
		})
	}
	
	// Add host reserved ports
	for _, v := range data.hostReservedPorts {
		allUrls = append(allUrls, generated.ServiceUrl{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	
	// Add service ports
	for _, v := range data.servicePorts {
		allUrls = append(allUrls, generated.ServiceUrl{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	
	// Add standard URLs
	if len(standardURLs) > 0 {
		for _, url := range standardURLs {
			allUrls = append(allUrls, generated.ServiceUrl{
				Service: url.Service,
				Url:     url.Url,
				Icon:    url.Icon,
				Fetched: false,
			})
		}
	}

	if print {
		s.prettyPrint(data, standardURLs)
	}

	return makeUnique(allUrls), nil
}

// ExtractURLs extracts service URLs from Nomad allocations
func (s *NomadService) ExtractURLs() ([]generated.ServiceUrl, error) {
	data, err := s.processAllocationsData()
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	standardURLs := s.standardURLs
	s.mu.RUnlock()

	// Combine service URLs with standard URLs
	result := make([]generated.ServiceUrl, 0, len(data.serviceUrls)+len(standardURLs))
	result = append(result, data.serviceUrls...)
	result = append(result, standardURLs...)

	return makeUnique(result), nil
}

// ExtractHostPorts extracts host reserved ports from Nomad allocations
func (s *NomadService) ExtractHostPorts() ([]generated.ServiceUrl, error) {
	data, err := s.processAllocationsData()
	if err != nil {
		return nil, err
	}

	return makeUnique(data.hostReservedPorts), nil
}

// ExtractServicePorts extracts service ports from Nomad allocations
func (s *NomadService) ExtractServicePorts() ([]generated.ServiceUrl, error) {
	data, err := s.processAllocationsData()
	if err != nil {
		return nil, err
	}

	return makeUnique(data.servicePorts), nil
}

// GetServiceStatus gets the status of a specific service
func (s *NomadService) GetServiceStatus(serviceName string) (map[string]string, error) {
	data, err := s.processAllocationsData()
	if err != nil {
		return nil, err
	}

	// Check service URLs
	for _, url := range data.serviceUrls {
		if url.Service == serviceName {
			return map[string]string{serviceName: url.Url}, nil
		}
	}
	
	// Check host reserved ports
	for _, port := range data.hostReservedPorts {
		if port.Service == serviceName {
			return map[string]string{serviceName: port.Url}, nil
		}
	}
	
	// Check service ports
	for _, port := range data.servicePorts {
		if port.Service == serviceName {
			return map[string]string{serviceName: port.Url}, nil
		}
	}

	return nil, nil
}

func (s *NomadService) RestartServiceAllocations(serviceName string) error {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return err
	}

	for _, allocation := range allocations {
		job, _, err := s.nomadClient.Jobs().Info(allocation.JobID, nil)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to get job info")
			return err
		}

		if *job.Name == serviceName {
			allocationInfo, _, err := s.nomadClient.Allocations().Info(allocation.ID, nil)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to get allocation info")
				return err
			}
			err = s.nomadClient.Allocations().Restart(allocationInfo, "", nil)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to restart service allocations")
				return err
			}
		}
	}

	return nil
}

// extractJobServices extracts all services from a job
func (s *NomadService) extractJobServices(job *api.Job) []*api.Service {
	services := []*api.Service{}
	for _, group := range job.TaskGroups {
		services = append(services, group.Services...)
		for _, task := range group.Tasks {
			services = append(services, task.Services...)
		}
	}
	return services
}

// processServiceTags processes service tags for URL and icon extraction
func (s *NomadService) processServiceTags(jobName string, services []*api.Service, data *allocationData) {
	for _, service := range services {
		s.getUrlDataFromTags(jobName, service.Name, service.Tags, data)
		s.getIconFromTags(jobName, service.Name, service.Tags, data)
	}
}

// processAllocation processes a single allocation and extracts relevant data
func (s *NomadService) processAllocation(allocation *api.AllocationListStub, data *allocationData) {
	allocationInfo, _, err := s.nomadClient.Allocations().Info(allocation.ID, nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to get allocation info")
		return
	}

	node, _, err := s.nomadClient.Nodes().Info(allocation.NodeID, nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to get node info")
		return
	}

	job, _, err := s.nomadClient.Jobs().Info(allocation.JobID, nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to get job info")
		return
	}

	// Process task groups for network ports
	for _, taskGroup := range job.TaskGroups {
		s.processTaskGroup(taskGroup, allocationInfo, node, job, data)
	}

	// Extract and process services from job
	services := s.extractJobServices(job)
	s.processServiceTags(*job.Name, services, data)
}

// processDynamicPorts processes dynamic ports for a task group
func (s *NomadService) processDynamicPorts(taskGroup *api.TaskGroup, allocation *api.Allocation, node *api.Node, job *api.Job, data *allocationData) {
	if len(taskGroup.Networks[0].DynamicPorts) == 0 {
		return
	}

	for _, port := range taskGroup.Networks[0].DynamicPorts {
		if port.To != 0 {
			for _, allocPort := range allocation.Resources.Networks[0].DynamicPorts {
				if allocPort.Label == port.Label {
					data.servicePorts = append(data.servicePorts, generated.ServiceUrl{
						Service: *job.Name + "-" + port.Label,
						Url:     fmt.Sprintf("%s:%d", node.HTTPAddr[:len(node.HTTPAddr)-5], allocPort.Value),
						Fetched: true,
					})
					break
				}
			}
		}
	}
}

// processReservedPorts processes reserved ports for a task group
func (s *NomadService) processReservedPorts(taskGroup *api.TaskGroup, node *api.Node, job *api.Job, data *allocationData) {
	if len(taskGroup.Networks[0].ReservedPorts) == 0 {
		return
	}

	for _, port := range taskGroup.Networks[0].ReservedPorts {
		data.hostReservedPorts = append(data.hostReservedPorts, generated.ServiceUrl{
			Service: *job.Name + "-" + port.Label,
			Url:     fmt.Sprintf("%s:%d", node.HTTPAddr[:len(node.HTTPAddr)-5], port.Value),
			Fetched: true,
		})
	}
}

// processTaskGroup processes a task group and extracts port information
func (s *NomadService) processTaskGroup(taskGroup *api.TaskGroup, allocation *api.Allocation, node *api.Node, job *api.Job, data *allocationData) {
	if len(taskGroup.Networks) == 0 {
		return
	}

	s.processDynamicPorts(taskGroup, allocation, node, job, data)
	s.processReservedPorts(taskGroup, node, job, data)
}

// extractURLFromTraefikTag extracts the URL from a Traefik rule tag
func (s *NomadService) extractURLFromTraefikTag(tag string) string {
	// Extract URL from format: traefik.http.routers.*.rule=Host(`example.com`)
	urlPart := strings.Split(tag, "(")[1]
	urlPart = strings.Split(urlPart, ")")[0]
	return urlPart[1 : len(urlPart)-1] // Remove surrounding backticks
}

// buildServiceName creates a service name based on job and task names
func (s *NomadService) buildServiceName(jobName, taskName string) string {
	if jobName == taskName {
		return jobName
	}
	return jobName + "-" + taskName
}

// addServiceURL adds a service URL to the data
func (s *NomadService) addServiceURL(serviceName, url string, data *allocationData) {
	data.serviceUrls = append(data.serviceUrls, generated.ServiceUrl{
		Service: serviceName,
		Url:     "https://" + url,
		Fetched: true,
	})
}

// getUrlDataFromTags extracts URL data from service tags
func (s *NomadService) getUrlDataFromTags(jobName string, taskName string, tags []string, data *allocationData) {
	if !slices.Contains(tags, "traefik.enable=true") {
		return
	}

	for _, tag := range tags {
		if traefikRuleTagRegex.MatchString(tag) {
			url := s.extractURLFromTraefikTag(tag)
			serviceName := s.buildServiceName(jobName, taskName)
			s.addServiceURL(serviceName, url, data)
			return
		}
	}
}

// extractIconFromTag extracts icon value from a tag using regex
func (s *NomadService) extractIconFromTag(tag string) string {
	matches := iconTagRegex.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// updateServiceIcon updates the icon for a service in the data
func (s *NomadService) updateServiceIcon(serviceName, icon string, data *allocationData) {
	for i, url := range data.serviceUrls {
		if url.Service == serviceName {
			data.serviceUrls[i].Icon = icon
			return
		}
	}
}

// getIconFromTags extracts icon data from service tags
func (s *NomadService) getIconFromTags(jobName string, taskName string, tags []string, data *allocationData) {
	for _, tag := range tags {
		if iconTagRegex.MatchString(tag) {
			icon := s.extractIconFromTag(tag)
			if icon != "" {
				serviceName := s.buildServiceName(jobName, taskName)
				s.updateServiceIcon(serviceName, icon, data)
				return
			}
		}
	}
}

// prettyPrint prints formatted tables of the extracted data
func (s *NomadService) prettyPrint(data *allocationData, standardURLs []generated.ServiceUrl) {
	s.printTable("Service", "URL", data.serviceUrls)
	s.printTable("Host Reserved Ports", "Port", data.hostReservedPorts)
	s.printTable("Service Ports", "Port", data.servicePorts)
	if len(standardURLs) > 0 {
		s.printTable("Standard URLs", "URL", standardURLs)
	}
}

// printTable prints a formatted table for the given data
func (s *NomadService) printTable(header1, header2 string, data []generated.ServiceUrl) {
	keys := make([]string, 0, len(data))
	for _, k := range data {
		keys = append(keys, k.Service)
	}
	sort.Strings(keys)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{header1, header2})

	for key := range keys {
		t.AppendRow([]interface{}{key, data[key].Url})
		if data[key].Icon != "" {
			t.AppendRow([]interface{}{"", fmt.Sprintf("Icon: %s", data[key].Icon)})
		}
	}

	t.Render()
}

func makeUnique(urls []generated.ServiceUrl) []generated.ServiceUrl {
	uniqueUrls := make(map[string]generated.ServiceUrl)
	for _, url := range urls {
		if _, exists := uniqueUrls[url.Service]; !exists {
			uniqueUrls[url.Service] = url
		}
	}

	result := make([]generated.ServiceUrl, 0, len(uniqueUrls))
	for _, url := range uniqueUrls {
		result = append(result, url)
	}

	// sort the result by service name alphabetically
	slices.SortFunc(result, func(a, b generated.ServiceUrl) int {
		return strings.Compare(a.Service, b.Service)
	})

	return result
}
