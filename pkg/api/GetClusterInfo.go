/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	"context"
	"fmt"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterInfo struct {
	NodesSize int
	AllocatableCPU,
	AllocatableMemory,
	AllocatablePods,
	AllocatableStorage,
	AllocatableStorageEphemeral float64
	PVSizeRWO    int
	PVSizeRWX    int
	PVStorageRWO float64
	PVStorageRWX float64
}

type ClusterInfoHuman struct {
	TotalNodes,
	StorageDisks,
	NodesMaxPods int
	StorageSize,
	NodesCPU,
	NodesMemory string
}

func (c *ClusterInfo) ToHuman() *ClusterInfoHuman {
	return &ClusterInfoHuman{
		TotalNodes:   c.NodesSize,
		StorageDisks: c.PVSizeRWO,
		StorageSize:  byteCountSI(c.PVStorageRWO),
		NodesCPU:     fmt.Sprintf("%.2f", c.AllocatableCPU),
		NodesMemory:  byteCountSI(c.AllocatableMemory),
		NodesMaxPods: int(c.AllocatablePods),
	}
}

func GetClusterInfo(ctx context.Context, name string) (*ClusterInfo, error) {
	clientset, err := client.GetClientset(name)
	if err != nil {
		return nil, errors.Wrap(err, "can get clientset")
	}

	result := ClusterInfo{}

	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "can not get nodes list")
	}

	result.NodesSize = len(nodeList.Items)

	for _, node := range nodeList.Items {
		result.AllocatableCPU += node.Status.Allocatable.Cpu().AsApproximateFloat64()
		result.AllocatableMemory += node.Status.Allocatable.Memory().AsApproximateFloat64()
		result.AllocatablePods += node.Status.Allocatable.Pods().AsApproximateFloat64()
		result.AllocatableStorage += node.Status.Allocatable.Storage().AsApproximateFloat64()
		result.AllocatableStorageEphemeral += node.Status.Allocatable.StorageEphemeral().AsApproximateFloat64()
	}

	pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "can not get pv list")
	}

	for _, pv := range pvList.Items {
		if pv.Status.Phase != "Bound" {
			continue
		}

		readWriteOnce := false

		for _, accessMode := range pv.Spec.AccessModes {
			if accessMode == corev1.ReadWriteOnce {
				readWriteOnce = true
			}
		}

		if readWriteOnce {
			result.PVSizeRWO++
			result.PVStorageRWO += pv.Spec.Capacity.Storage().AsApproximateFloat64()
		} else {
			result.PVSizeRWX++
			result.PVStorageRWX += pv.Spec.Capacity.Storage().AsApproximateFloat64()
		}
	}

	return &result, nil
}
