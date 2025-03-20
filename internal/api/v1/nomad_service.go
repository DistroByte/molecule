package v1

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/nomad/api"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
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
}

var (
	standardURLs      = make(map[string]string)
	serviceUrls       = make(map[string]string)
	hostReservedPorts = make(map[string]string)
	servicePorts      = make(map[string]string)
)

func NewNomadService(nomadClient *api.Client) NomadServiceInterface {
	standardURLs["nomad"] = "http://zeus.internal:4646"
	standardURLs["consul"] = "http://zeus.internal:8500"
	standardURLs["traefik"] = "http://zeus.internal:8081"
	standardURLs["plausible"] = "https://plausible.dbyte.xyz"

	return &NomadService{nomadClient: nomadClient}
}

func (s *NomadService) ExtractAll(print bool) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list allocations")
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
		log.Error().Err(err).Msg("Failed to list allocations")
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
		log.Error().Err(err).Msg("Failed to list allocations")
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
		log.Error().Err(err).Msg("Failed to list allocations")
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

func (s *NomadService) processAllocation(allocation *api.AllocationListStub) {
	services := []*api.Service{}

	allocationInfo, _, err := s.nomadClient.Allocations().Info(allocation.ID, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get allocation info")
		return
	}

	node, _, err := s.nomadClient.Nodes().Info(allocation.NodeID, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get node info")
		return
	}

	job, _, err := s.nomadClient.Jobs().Info(allocation.JobID, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get job info")
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
