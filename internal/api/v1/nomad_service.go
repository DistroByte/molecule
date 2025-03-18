package v1

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/jedib0t/go-pretty/v6/table"
)

type NomadService struct {
	nomadClient *api.Client
}

var (
	serviceUrls       = make(map[string]string)
	hostReservedPorts = make(map[string]string)
	servicePorts      = make(map[string]string)
)

func NewNomadService(nomadClient *api.Client) *NomadService {
	return &NomadService{nomadClient: nomadClient}
}

func (s *NomadService) ExtractAll(print bool) (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

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

	if print {
		prettyPrint()
	}

	return allUrls, nil
}

func (s *NomadService) ExtractURLs() (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	return serviceUrls, nil
}

func (s *NomadService) ExtractHostPorts() (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	return hostReservedPorts, nil
}

func (s *NomadService) ExtractServicePorts() (map[string]string, error) {
	allocations, _, err := s.nomadClient.Allocations().List(nil)
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocations {
		s.processAllocation(allocation)
	}

	return servicePorts, nil
}

func (s *NomadService) processAllocation(allocation *api.AllocationListStub) {
	services := []*api.Service{}

	allocationInfo, _, err := s.nomadClient.Allocations().Info(allocation.ID, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	node, _, err := s.nomadClient.Nodes().Info(allocation.NodeID, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	job, _, err := s.nomadClient.Jobs().Info(allocation.JobID, nil)
	if err != nil {
		fmt.Println(err)
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
