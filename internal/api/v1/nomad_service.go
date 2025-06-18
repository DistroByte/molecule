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

type NomadService struct {
	nomadClient *api.Client
	mu          sync.Mutex
}

type NomadServiceInterface interface {
	ExtractAll(print bool) ([]generated.GetUrls200ResponseInner, error)
	ExtractURLs() ([]generated.GetUrls200ResponseInner, error)
	ExtractHostPorts() ([]generated.GetUrls200ResponseInner, error)
	ExtractServicePorts() ([]generated.GetUrls200ResponseInner, error)
	GetServiceStatus(service string) (map[string]string, error)
	RestartServiceAllocations(service string) error
}

var (
	standardURLs      = []generated.GetUrls200ResponseInner{}
	serviceUrls       = []generated.GetUrls200ResponseInner{}
	hostReservedPorts = []generated.GetUrls200ResponseInner{}
	servicePorts      = []generated.GetUrls200ResponseInner{}
)

func NewNomadService(nomadClient *api.Client, staticUrls []generated.GetUrls200ResponseInner) NomadServiceInterface {
	standardURLs = staticUrls

	return &NomadService{nomadClient: nomadClient}
}

func (s *NomadService) ExtractAll(print bool) ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	serviceUrls = []generated.GetUrls200ResponseInner{}
	hostReservedPorts = []generated.GetUrls200ResponseInner{}
	servicePorts = []generated.GetUrls200ResponseInner{}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	allUrls := []generated.GetUrls200ResponseInner{}
	for _, v := range serviceUrls {
		allUrls = append(allUrls, generated.GetUrls200ResponseInner{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	for _, v := range hostReservedPorts {
		allUrls = append(allUrls, generated.GetUrls200ResponseInner{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	for _, v := range servicePorts {
		allUrls = append(allUrls, generated.GetUrls200ResponseInner{
			Service: v.Service,
			Url:     v.Url,
			Fetched: true,
		})
	}
	if len(standardURLs) > 0 {
		for _, url := range standardURLs {
			allUrls = append(allUrls, generated.GetUrls200ResponseInner{
				Service: url.Service,
				Url:     url.Url,
				Fetched: false,
			})
		}
	}

	if print {
		prettyPrint()
	}

	return makeUnique(allUrls), nil
}

func (s *NomadService) ExtractURLs() ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	serviceUrls = []generated.GetUrls200ResponseInner{}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	serviceUrls = append(makeUnique(serviceUrls), standardURLs...)

	return serviceUrls, nil
}

func (s *NomadService) ExtractHostPorts() ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	hostReservedPorts = []generated.GetUrls200ResponseInner{}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return makeUnique(hostReservedPorts), nil
}

func (s *NomadService) ExtractServicePorts() ([]generated.GetUrls200ResponseInner, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	servicePorts = []generated.GetUrls200ResponseInner{}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return makeUnique(servicePorts), nil
}

func (s *NomadService) GetServiceStatus(serviceName string) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to list allocations")
		return nil, err
	}

	serviceUrls = []generated.GetUrls200ResponseInner{}
	hostReservedPorts = []generated.GetUrls200ResponseInner{}
	servicePorts = []generated.GetUrls200ResponseInner{}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, url := range serviceUrls {
		if url.Service == serviceName {
			return map[string]string{serviceName: url.Url}, nil
		}
	}
	for _, port := range hostReservedPorts {
		if port.Service == serviceName {
			return map[string]string{serviceName: port.Url}, nil
		}
	}
	for _, port := range servicePorts {
		if port.Service == serviceName {
			return map[string]string{serviceName: port.Url}, nil
		}
	}

	return nil, nil
}

// when called, this function will restart all allocations for a specific service
// this is useful when a new version of a service is built and you want to restart all instances
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
						servicePorts = append(servicePorts, generated.GetUrls200ResponseInner{
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
			hostReservedPorts = append(hostReservedPorts, generated.GetUrls200ResponseInner{
				Service: *job.Name + "-" + port.Label,
				Url:     fmt.Sprintf("%s:%d", node.HTTPAddr[:len(node.HTTPAddr)-5], port.Value),
				Fetched: true,
			})
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
					serviceUrls = append(serviceUrls, generated.GetUrls200ResponseInner{
						Service: jobName,
						Url:     "https://" + url,
						Fetched: true,
					})

					return
				} else {
					serviceUrls = append(serviceUrls, generated.GetUrls200ResponseInner{
						Service: jobName + "-" + taskName,
						Url:     "https://" + url,
						Fetched: true,
					})

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

func printTable(header1, header2 string, data []generated.GetUrls200ResponseInner) {
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].Service < result[j].Service
	})
	return result
}
