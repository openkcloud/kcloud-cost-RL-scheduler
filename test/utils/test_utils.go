/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kcloudv1alpha1 "github.com/KETI-Cloud-Platform/k8s-workload-operator/api/v1alpha1"
)

// CreateTestWorkloadOptimizer creates a test WorkloadOptimizer resource
func CreateTestWorkloadOptimizer(name, namespace string) *kcloudv1alpha1.WorkloadOptimizer {
	return &kcloudv1alpha1.WorkloadOptimizer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kcloudv1alpha1.WorkloadOptimizerSpec{
			WorkloadType: "training",
			Priority:     5,
			ResourceRequirements: kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("2"),
				Memory: resource.MustParse("4Gi"),
				GPU:    1,
				NPU:    0,
			},
			CostConstraints: kcloudv1alpha1.CostConstraints{
				MaxCostPerHour: 10.0,
				BudgetLimit:    1000.0,
			},
			PowerConstraints: kcloudv1alpha1.PowerConstraints{
				MaxPowerUsage: 500.0,
			},
		},
	}
}

// CreateTestNode creates a test Node resource
func CreateTestNode(name string, nodeType string) *corev1.Node {
	labels := map[string]string{
		"node-type": nodeType,
	}

	var allocatable corev1.ResourceList
	switch nodeType {
	case "cpu-optimized":
		allocatable = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
			corev1.ResourceGPU:    resource.MustParse("0"),
		}
	case "gpu-optimized":
		allocatable = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("8"),
			corev1.ResourceMemory: resource.MustParse("16Gi"),
			corev1.ResourceGPU:    resource.MustParse("2"),
		}
	case "npu-optimized":
		allocatable = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("6"),
			corev1.ResourceMemory: resource.MustParse("12Gi"),
			corev1.ResourceNPU:    resource.MustParse("1"),
		}
	default:
		allocatable = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
			corev1.ResourceGPU:    resource.MustParse("0"),
		}
	}

	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Status: corev1.NodeStatus{
			Allocatable: allocatable,
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}

// CreateTestPod creates a test Pod resource
func CreateTestPod(name, namespace string, workloadName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"workloadoptimizer.kcloud.io/name": workloadName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
		},
	}
}

// CreateTestCostPolicy creates a test CostPolicy resource
func CreateTestCostPolicy(name, namespace string) *kcloudv1alpha1.CostPolicy {
	return &kcloudv1alpha1.CostPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kcloudv1alpha1.CostPolicySpec{
			BudgetLimit:      1000.0,
			CostPerHourLimit: 10.0,
			SpotInstancePolicy: kcloudv1alpha1.SpotInstancePolicy{
				Enabled:  true,
				MaxPrice: 5.0,
			},
			AlertThresholds: kcloudv1alpha1.CostAlertThresholds{
				BudgetUtilization: 80.0,
				CostIncrease:      20.0,
			},
		},
	}
}

// CreateTestPowerPolicy creates a test PowerPolicy resource
func CreateTestPowerPolicy(name, namespace string) *kcloudv1alpha1.PowerPolicy {
	return &kcloudv1alpha1.PowerPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kcloudv1alpha1.PowerPolicySpec{
			MaxPowerUsage:    500.0,
			EfficiencyTarget: 80.0,
			GreenEnergyPolicy: kcloudv1alpha1.GreenEnergyPolicy{
				Enabled:    true,
				Preference: "renewable",
			},
			AlertThresholds: kcloudv1alpha1.PowerAlertThresholds{
				PowerUsage: 90.0,
				Efficiency: 70.0,
			},
		},
	}
}

// CreateTestWorkloadState creates a test WorkloadState
func CreateTestWorkloadState(workload *kcloudv1alpha1.WorkloadOptimizer, pods []corev1.Pod, nodes []corev1.Node) *WorkloadState {
	return &WorkloadState{
		Workload: workload,
		Pods:     pods,
		Nodes:    nodes,
	}
}

// CreateTestCurrentState creates a test CurrentState
func CreateTestCurrentState(pods []corev1.Pod, nodes []corev1.Node) *CurrentState {
	return &CurrentState{
		Pods:  pods,
		Nodes: nodes,
	}
}

// CreateTestResourceRequirements creates test resource requirements
func CreateTestResourceRequirements(cpu, memory, gpu, npu string) kcloudv1alpha1.ResourceRequirements {
	return kcloudv1alpha1.ResourceRequirements{
		CPU:    resource.MustParse(cpu),
		Memory: resource.MustParse(memory),
		GPU:    resource.MustParse(gpu).Value(),
		NPU:    resource.MustParse(npu).Value(),
	}
}

// CreateTestCostConstraints creates test cost constraints
func CreateTestCostConstraints(maxCostPerHour, budgetLimit float64) kcloudv1alpha1.CostConstraints {
	return kcloudv1alpha1.CostConstraints{
		MaxCostPerHour: maxCostPerHour,
		BudgetLimit:    budgetLimit,
	}
}

// CreateTestPowerConstraints creates test power constraints
func CreateTestPowerConstraints(maxPowerUsage float64) kcloudv1alpha1.PowerConstraints {
	return kcloudv1alpha1.PowerConstraints{
		MaxPowerUsage: maxPowerUsage,
	}
}

// CreateTestPlacementPolicy creates test placement policy
func CreateTestPlacementPolicy(nodeAffinity *corev1.NodeAffinity) *kcloudv1alpha1.PlacementPolicy {
	return &kcloudv1alpha1.PlacementPolicy{
		NodeAffinity: nodeAffinity,
	}
}

// CreateTestAutoScalingSpec creates test auto-scaling spec
func CreateTestAutoScalingSpec(enabled bool, minReplicas, maxReplicas int32, targetCPU, targetMemory int32) *kcloudv1alpha1.AutoScalingSpec {
	return &kcloudv1alpha1.AutoScalingSpec{
		Enabled:      enabled,
		MinReplicas:  minReplicas,
		MaxReplicas:  maxReplicas,
		TargetCPU:    targetCPU,
		TargetMemory: targetMemory,
	}
}
