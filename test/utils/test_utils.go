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
	"bytes"
	"os/exec"
	"strings"

	"go.yaml.in/yaml/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

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

// YAMLToReader converts a Kubernetes object to YAML and returns a reader
func YAMLToReader(obj runtime.Object) *bytes.Reader {
	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(yamlBytes)
}

// Run executes a command and returns the output
func Run(cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// GetNonEmptyLines splits output into lines and returns non-empty ones
func GetNonEmptyLines(output string) []string {
	lines := strings.Split(output, "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, strings.TrimSpace(line))
		}
	}
	return result
}

// WaitForResource waits for a resource to reach a specific condition
func WaitForResource(resourceType, name, namespace, condition string, timeout int) error {
	cmd := exec.Command("kubectl", "wait", "--for=condition="+condition,
		resourceType+"/"+name, "-n", namespace, "--timeout="+string(rune(timeout))+"s")
	_, err := Run(cmd)
	return err
}

// GetResourceField gets a specific field from a Kubernetes resource
func GetResourceField(resourceType, name, namespace, field string) (string, error) {
	cmd := exec.Command("kubectl", "get", resourceType, name, "-n", namespace,
		"-o", "jsonpath={."+field+"}")
	return Run(cmd)
}

// ApplyResource applies a Kubernetes resource from YAML
func ApplyResource(yamlContent string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(yamlContent)
	_, err := Run(cmd)
	return err
}

// DeleteResource deletes a Kubernetes resource
func DeleteResource(resourceType, name, namespace string) error {
	cmd := exec.Command("kubectl", "delete", resourceType, name, "-n", namespace, "--ignore-not-found=true")
	_, err := Run(cmd)
	return err
}

// CreateNamespace creates a Kubernetes namespace
func CreateNamespace(name string) error {
	cmd := exec.Command("kubectl", "create", "namespace", name)
	_, err := Run(cmd)
	return err
}

// DeleteNamespace deletes a Kubernetes namespace
func DeleteNamespace(name string) error {
	cmd := exec.Command("kubectl", "delete", "namespace", name, "--ignore-not-found=true")
	_, err := Run(cmd)
	return err
}

// LabelNamespace adds labels to a namespace
func LabelNamespace(name string, labels map[string]string) error {
	var labelArgs []string
	for key, value := range labels {
		labelArgs = append(labelArgs, key+"="+value)
	}
	cmd := exec.Command("kubectl", "label", "namespace", name, strings.Join(labelArgs, ","), "--overwrite")
	_, err := Run(cmd)
	return err
}

// GetPodsByLabel gets pods by label selector
func GetPodsByLabel(namespace, labelSelector string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-l", labelSelector, "-o", "name")
	output, err := Run(cmd)
	if err != nil {
		return nil, err
	}
	return GetNonEmptyLines(output), nil
}

// GetEvents gets events for a namespace
func GetEvents(namespace string) (string, error) {
	cmd := exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
	return Run(cmd)
}

// GetLogs gets logs from a pod
func GetLogs(podName, namespace string) (string, error) {
	cmd := exec.Command("kubectl", "logs", podName, "-n", namespace)
	return Run(cmd)
}

// DescribeResource describes a Kubernetes resource
func DescribeResource(resourceType, name, namespace string) (string, error) {
	cmd := exec.Command("kubectl", "describe", resourceType, name, "-n", namespace)
	return Run(cmd)
}

// CheckResourceExists checks if a resource exists
func CheckResourceExists(resourceType, name, namespace string) bool {
	cmd := exec.Command("kubectl", "get", resourceType, name, "-n", namespace)
	_, err := Run(cmd)
	return err == nil
}

// GetResourceStatus gets the status of a resource
func GetResourceStatus(resourceType, name, namespace string) (string, error) {
	return GetResourceField(resourceType, name, namespace, "status")
}

// WaitForWorkloadOptimizerPhase waits for WorkloadOptimizer to reach a specific phase
func WaitForWorkloadOptimizerPhase(name, namespace, phase string, timeout int) error {
	return WaitForResource("workloadoptimizer", name, namespace, "phase="+phase, timeout)
}

// GetWorkloadOptimizerPhase gets the current phase of a WorkloadOptimizer
func GetWorkloadOptimizerPhase(name, namespace string) (string, error) {
	return GetResourceField("workloadoptimizer", name, namespace, "status.phase")
}

// GetWorkloadOptimizerScore gets the optimization score of a WorkloadOptimizer
func GetWorkloadOptimizerScore(name, namespace string) (string, error) {
	return GetResourceField("workloadoptimizer", name, namespace, "status.optimizationScore")
}

// CreateTestDeployment creates a test deployment
func CreateTestDeployment(name, namespace string) *corev1.Deployment {
	return &corev1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// int32Ptr returns a pointer to an int32
func int32Ptr(i int32) *int32 { return &i }
