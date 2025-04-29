package v1

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/DistroByte/molecule/logger"
	"github.com/hashicorp/nomad/api"
	"github.com/jedib0t/go-pretty/v6/table"
)

type NomadService struct {
	nomadClient *api.Client
	mu          sync.Mutex
}

type NomadServiceInterface interface {
	ExtractAll(print bool) (map[string]string, error)
	ExtractURLs() (map[string]string, error)
	ExtractHostPorts() (map[string]string, error)
	ExtractServicePorts() (map[string]string, error)
	GetServiceStatus(service string) (map[string]string, error)
	RestartServiceAllocations(service string) (map[string]string, error)
}

var (
	standardURLs      = make(map[string]string)
	serviceUrls       = make(map[string]string)
	hostReservedPorts = make(map[string]string)
	servicePorts      = make(map[string]string)
)

func NewNomadService(nomadClient *api.Client, staticUrls map[string]string) NomadServiceInterface {
	for key, value := range staticUrls {
		standardURLs[key] = value
	}

	return &NomadService{nomadClient: nomadClient}
}

func (s *NomadService) ExtractAll(print bool) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	serviceUrls = make(map[string]string)
	hostReservedPorts = make(map[string]string)
	servicePorts = make(map[string]string)

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	allUrls := make(map[string]string)
	for k, v := range serviceUrls {
		allUrls[k] = v
	}
	for k, v := range hostReservedPorts {
		allUrls[k] = v
	}
	for k, v := range servicePorts {
		allUrls[k] = v
	}
	for k, v := range standardURLs {
		allUrls[k] = v
	}

	if print {
		prettyPrint()
	}

	return allUrls, nil
}

func (s *NomadService) ExtractURLs() (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	serviceUrls = make(map[string]string)

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range standardURLs {
		serviceUrls[k] = v
	}

	return serviceUrls, nil
}

func (s *NomadService) ExtractHostPorts() (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	hostReservedPorts = make(map[string]string)

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return hostReservedPorts, nil
}

func (s *NomadService) ExtractServicePorts() (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	servicePorts = make(map[string]string)

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return servicePorts, nil
}

func (s *NomadService) GetServiceStatus(serviceName string) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	serviceUrls = make(map[string]string)
	hostReservedPorts = make(map[string]string)
	servicePorts = make(map[string]string)

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := serviceUrls[serviceName]; ok {
		return map[string]string{serviceName: val}, nil
	}

	if val, ok := hostReservedPorts[serviceName]; ok {
		return map[string]string{serviceName: val}, nil
	}

	if val, ok := servicePorts[serviceName]; ok {
		return map[string]string{serviceName: val}, nil
	}

	return nil, nil
}

// when called, this function will restart all allocations for a specific service
// this is useful when a new version of a service is built and you want to restart all instances
func (s *NomadService) RestartServiceAllocations(serviceName string) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	for _, allocation := range allocations {
		job, _, err := s.nomadClient.Jobs().Info(allocation.JobID, nil)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to get job info")
			return nil, err
		}

		if *job.Name == serviceName {
			allocationInfo, _, err := s.nomadClient.Allocations().Info(allocation.ID, nil)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to get allocation info")
				return nil, err
			}
			err = s.nomadClient.Allocations().Restart(allocationInfo, "", nil)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to restart service allocations")
				return nil, err
			}
		}
	}

	return nil, nil
}

func (s *NomadService) processAllocation(allocation *api.AllocationListStub) {
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
		s.processTaskGroup(taskGroup, allocationInfo, node, job)
	}

	for _, group := range job.TaskGroups {
		services = append(services, group.Services...)

		for _, task := range group.Tasks {
			services = append(services, task.Services...)
		}
	}

	for _, service := range services {
		s.getTraefikURL(*job.Name, service.Name, service.Tags)
	}
}

func (s *NomadService) processTaskGroup(taskGroup *api.TaskGroup, allocation *api.Allocation, node *api.Node, job *api.Job) {
	if len(taskGroup.Networks) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(taskGroup.Networks[0].DynamicPorts) != 0 {
		for _, port := range taskGroup.Networks[0].DynamicPorts {
			if port.To != 0 {
				for _, allocPort := range allocation.Resources.Networks[0].DynamicPorts {
					if allocPort.Label == port.Label {
						servicePorts[*job.Name+"-"+port.Label] = fmt.Sprintf("%s:%d", node.HTTPAddr[:len(node.HTTPAddr)-5], allocPort.Value)
						break
					}
				}
			}
		}
	}

	if len(taskGroup.Networks[0].ReservedPorts) != 0 {
		for _, port := range taskGroup.Networks[0].ReservedPorts {
			hostReservedPorts[*job.Name+"-"+port.Label] = fmt.Sprintf("%s:%d", node.HTTPAddr[:len(node.HTTPAddr)-5], port.Value)
		}
	}
}

func (s *NomadService) getTraefikURL(jobName string, taskName string, tags []string) {
	if slices.Contains(tags, "traefik.enable=true") {
		re := regexp.MustCompile("traefik.http.routers.*.rule")
		for _, tag := range tags {
			if re.MatchString(tag) {
				url := strings.Split(tag, "(")[1]
				url = strings.Split(url, ")")[0]
				url = url[1 : len(url)-1]

				s.mu.Lock()
				defer s.mu.Unlock()

				if jobName == taskName {
					serviceUrls[jobName] = "https://" + url
					return
				} else {
					serviceUrls[jobName+"-"+taskName] = "https://" + url
					return
				}
			}
		}
	}
}

func prettyPrint() {
	printTable("Service", "URL", serviceUrls)
	printTable("Host Reserved Ports", "Port", hostReservedPorts)
	printTable("Service Ports", "Port", servicePorts)
}

func printTable(header1, header2 string, data map[string]string) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{header1, header2})

	for _, key := range keys {
		t.AppendRow([]interface{}{key, data[key]})
	}

	t.Render()
}
