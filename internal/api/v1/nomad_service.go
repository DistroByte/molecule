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
	standardURLs  []generated.GetUrls200ResponseInner
	mu            sync.RWMutex
}

// NomadServiceInterface defines the interface for Nomad service operations
type NomadServiceInterface interface {
	ExtractAll(print bool) ([]generated.GetUrls200ResponseInner, error)
	ExtractURLs() ([]generated.GetUrls200ResponseInner, error)
	ExtractHostPorts() ([]generated.GetUrls200ResponseInner, error)
	ExtractServicePorts() ([]generated.GetUrls200ResponseInner, error)
	GetServiceStatus(service string) (map[string]string, error)
	RestartServiceAllocations(service string) error
}

var (
	traefikRuleTagRegex = regexp.MustCompile("traefik.http.routers.*.rule")
	iconTagRegex        = regexp.MustCompile("icon=(.*)")
)

// allocationData holds the extracted data from allocations
type allocationData struct {
	serviceUrls       []generated.GetUrls200ResponseInner
	hostReservedPorts []generated.GetUrls200ResponseInner
	servicePorts      []generated.GetUrls200ResponseInner
}

// NewNomadService creates a new NomadService instance
func NewNomadService(nomadClient *api.Client, staticUrls []generated.GetUrls200ResponseInner) NomadServiceInterface {
	return &NomadService{
		nomadClient:  nomadClient,
		standardURLs: staticUrls,
	}
}

// ExtractAll extracts all URLs from Nomad allocations
func (s *NomadService) ExtractAll(print bool) ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	data := &allocationData{
		serviceUrls:       []generated.GetUrls200ResponseInner{},
		hostReservedPorts: []generated.GetUrls200ResponseInner{},
		servicePorts:      []generated.GetUrls200ResponseInner{},
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation, data)
	}

	s.mu.RLock()
	standardURLs := s.standardURLs
	s.mu.RUnlock()

	allUrls := []generated.GetUrls200ResponseInner{}
	
	// Add service URLs
	for _, v := range data.serviceUrls {
		allUrls = append(allUrls, generated.GetUrls200ResponseInner{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
			Icon:    v.Icon,
		})
	}
	
	// Add host reserved ports
	for _, v := range data.hostReservedPorts {
		allUrls = append(allUrls, generated.GetUrls200ResponseInner{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	
	// Add service ports
	for _, v := range data.servicePorts {
		allUrls = append(allUrls, generated.GetUrls200ResponseInner{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	
	// Add standard URLs
	if len(standardURLs) > 0 {
		for _, url := range standardURLs {
			allUrls = append(allUrls, generated.GetUrls200ResponseInner{
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
func (s *NomadService) ExtractURLs() ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	data := &allocationData{
		serviceUrls:       []generated.GetUrls200ResponseInner{},
		hostReservedPorts: []generated.GetUrls200ResponseInner{},
		servicePorts:      []generated.GetUrls200ResponseInner{},
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation, data)
	}

	s.mu.RLock()
	standardURLs := s.standardURLs
	s.mu.RUnlock()

	// Combine service URLs with standard URLs
	result := make([]generated.GetUrls200ResponseInner, 0, len(data.serviceUrls)+len(standardURLs))
	result = append(result, data.serviceUrls...)
	result = append(result, standardURLs...)

	return makeUnique(result), nil
}

// ExtractHostPorts extracts host reserved ports from Nomad allocations
func (s *NomadService) ExtractHostPorts() ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	data := &allocationData{
		serviceUrls:       []generated.GetUrls200ResponseInner{},
		hostReservedPorts: []generated.GetUrls200ResponseInner{},
		servicePorts:      []generated.GetUrls200ResponseInner{},
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation, data)
	}

	return makeUnique(data.hostReservedPorts), nil
}

// ExtractServicePorts extracts service ports from Nomad allocations
func (s *NomadService) ExtractServicePorts() ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	data := &allocationData{
		serviceUrls:       []generated.GetUrls200ResponseInner{},
		hostReservedPorts: []generated.GetUrls200ResponseInner{},
		servicePorts:      []generated.GetUrls200ResponseInner{},
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation, data)
	}

	return makeUnique(data.servicePorts), nil
}

// GetServiceStatus gets the status of a specific service
func (s *NomadService) GetServiceStatus(serviceName string) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	data := &allocationData{
		serviceUrls:       []generated.GetUrls200ResponseInner{},
		hostReservedPorts: []generated.GetUrls200ResponseInner{},
		servicePorts:      []generated.GetUrls200ResponseInner{},
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation, data)
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

// processAllocation processes a single allocation and extracts relevant data
func (s *NomadService) processAllocation(allocation *api.AllocationListStub, data *allocationData) {
	services := []*api.Service{}

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

	for _, taskGroup := range job.TaskGroups {
		s.processTaskGroup(taskGroup, allocationInfo, node, job, data)
	}

	for _, group := range job.TaskGroups {
		services = append(services, group.Services...)

		for _, task := range group.Tasks {
			services = append(services, task.Services...)
		}
	}

	for _, service := range services {
		s.getUrlDataFromTags(*job.Name, service.Name, service.Tags, data)
		s.getIconFromTags(*job.Name, service.Name, service.Tags, data)
	}
}

// processTaskGroup processes a task group and extracts port information
func (s *NomadService) processTaskGroup(taskGroup *api.TaskGroup, allocation *api.Allocation, node *api.Node, job *api.Job, data *allocationData) {
	if len(taskGroup.Networks) == 0 {
		return
	}

	if len(taskGroup.Networks[0].DynamicPorts) != 0 {
		for _, port := range taskGroup.Networks[0].DynamicPorts {
			if port.To != 0 {
				for _, allocPort := range allocation.Resources.Networks[0].DynamicPorts {
					if allocPort.Label == port.Label {
						data.servicePorts = append(data.servicePorts, generated.GetUrls200ResponseInner{
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

	if len(taskGroup.Networks[0].ReservedPorts) != 0 {
		for _, port := range taskGroup.Networks[0].ReservedPorts {
			data.hostReservedPorts = append(data.hostReservedPorts, generated.GetUrls200ResponseInner{
				Service: *job.Name + "-" + port.Label,
				Url:     fmt.Sprintf("%s:%d", node.HTTPAddr[:len(node.HTTPAddr)-5], port.Value),
				Fetched: true,
			})
		}
	}
}

// getUrlDataFromTags extracts URL data from service tags
func (s *NomadService) getUrlDataFromTags(jobName string, taskName string, tags []string, data *allocationData) {
	if slices.Contains(tags, "traefik.enable=true") {
		for _, tag := range tags {
			if traefikRuleTagRegex.MatchString(tag) {
				url := strings.Split(tag, "(")[1]
				url = strings.Split(url, ")")[0]
				url = url[1 : len(url)-1]

				if jobName == taskName {
					data.serviceUrls = append(data.serviceUrls, generated.GetUrls200ResponseInner{
						Service: jobName,
						Url:     "https://" + url,
						Fetched: true,
					})
				} else {
					data.serviceUrls = append(data.serviceUrls, generated.GetUrls200ResponseInner{
						Service: jobName + "-" + taskName,
						Url:     "https://" + url,
						Fetched: true,
					})
				}
				return
			}
		}
	}
}

// getIconFromTags extracts icon data from service tags
func (s *NomadService) getIconFromTags(jobName string, taskName string, tags []string, data *allocationData) {
	for _, tag := range tags {
		if iconTagRegex.MatchString(tag) {
			icon := iconTagRegex.FindStringSubmatch(tag)
			if len(icon) > 0 {
				for i, url := range data.serviceUrls {
					if url.Service == jobName || url.Service == jobName+"-"+taskName {
						data.serviceUrls[i].Icon = icon[1]
						return
					}
				}
			}
		}
	}
}

// prettyPrint prints formatted tables of the extracted data
func (s *NomadService) prettyPrint(data *allocationData, standardURLs []generated.GetUrls200ResponseInner) {
	s.printTable("Service", "URL", data.serviceUrls)
	s.printTable("Host Reserved Ports", "Port", data.hostReservedPorts)
	s.printTable("Service Ports", "Port", data.servicePorts)
	if len(standardURLs) > 0 {
		s.printTable("Standard URLs", "URL", standardURLs)
	}
}

// printTable prints a formatted table for the given data
func (s *NomadService) printTable(header1, header2 string, data []generated.GetUrls200ResponseInner) {
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

func makeUnique(urls []generated.GetUrls200ResponseInner) []generated.GetUrls200ResponseInner {
	uniqueUrls := make(map[string]generated.GetUrls200ResponseInner)
	for _, url := range urls {
		if _, exists := uniqueUrls[url.Service]; !exists {
			uniqueUrls[url.Service] = url
		}
	}

	result := make([]generated.GetUrls200ResponseInner, 0, len(uniqueUrls))
	for _, url := range uniqueUrls {
		result = append(result, url)
	}

	// sort the result by service name alphabetically
	slices.SortFunc(result, func(a, b generated.GetUrls200ResponseInner) int {
		return strings.Compare(a.Service, b.Service)
	})

	return result
}
