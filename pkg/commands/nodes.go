package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/nerzhul/kubectl-thunder/pkg/kubernetes"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type nodeResources struct {
	allocatedCPU    resource.Quantity
	allocatedMemory resource.Quantity

	allocatableCPU    resource.Quantity
	allocatableMemory resource.Quantity

	labels map[string]string
}

func Nodes_can_allocate(c *cli.Context) error {
	kClient, err := kubernetes.CreateClient()
	if err != nil {
		return err
	}

	memory := c.String("memory")
	cpu := c.String("cpu")
	labels := c.StringSlice("show-labels")

	log.Printf("Checking if nodes can allocate %s MiB of memory and %s millicores of CPU", memory, cpu)

	nrStore := make(map[string]*nodeResources, 0)
	{
		pods, err := kClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pod := range pods.Items {
			nodeName := pod.Spec.NodeName
			// ignore non allocated pods
			if nodeName == "" {
				continue
			}

			// ignore non running pods
			if pod.Status.Phase != v1.PodRunning {
				continue
			}

			cpuQuantity := resource.Quantity{}
			memQuantity := resource.Quantity{}

			for _, container := range pod.Spec.Containers {
				for resourceName, resourceValue := range container.Resources.Requests {
					switch resourceName {
					case v1.ResourceCPU:
						cpuQuantity.Add(resourceValue)
					case v1.ResourceMemory:
						memQuantity.Add(resourceValue)
					}
				}
			}

			if _, ok := nrStore[nodeName]; !ok {
				nrStore[nodeName] = &nodeResources{
					allocatedCPU:    cpuQuantity,
					allocatedMemory: memQuantity,
				}
			} else {
				nrStore[nodeName].allocatedCPU.Add(cpuQuantity)
				nrStore[nodeName].allocatedMemory.Add(memQuantity)
			}
		}
	}

	nodes, err := kClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		if _, ok := nrStore[node.Name]; !ok {
			nrStore[node.Name] = &nodeResources{
				allocatableCPU:    node.Status.Allocatable[v1.ResourceCPU],
				allocatableMemory: node.Status.Allocatable[v1.ResourceMemory],
			}
		} else {
			nrStore[node.Name].allocatableCPU = node.Status.Allocatable[v1.ResourceCPU]
			nrStore[node.Name].allocatableMemory = node.Status.Allocatable[v1.ResourceMemory]
		}

		nrStore[node.Name].labels = node.Labels
	}

	for nodeName, nodeResources := range nrStore {
		remainingMem := nodeResources.allocatableMemory
		remainingMem.Sub(nodeResources.allocatedMemory)

		remainingCPU := nodeResources.allocatableCPU
		remainingCPU.Sub(nodeResources.allocatedCPU)

		if remainingMem.Cmp(resource.MustParse(memory)) < 0 || remainingCPU.Cmp(resource.MustParse(cpu)) < 0 {
			logStr := fmt.Sprintf("Node %s allocation failure (", nodeName)
			if remainingMem.Cmp(resource.MustParse(memory)) < 0 {
				logStr += fmt.Sprintf("mem: %s>%s", memory, remainingMem.String())
			}

			if remainingCPU.Cmp(resource.MustParse(cpu)) < 0 {
				if remainingMem.Cmp(resource.MustParse(memory)) < 0 {
					logStr += ", "
				}
				logStr += fmt.Sprintf("cpu: %s>%s", cpu, remainingCPU.String())
			}

			logStr += ")"

			if len(labels) > 0 {
				for _, label := range labels {
					if val, ok := nodeResources.labels[label]; ok {
						logStr += fmt.Sprintf(" %s=%s", label, val)
					}
				}
			}

			log.Print(logStr)
		}
	}

	return nil
}
