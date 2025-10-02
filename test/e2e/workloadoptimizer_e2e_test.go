//go:build e2e
// +build e2e

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

package e2e

import (
	"context"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kcloudv1alpha1 "github.com/KETI-Cloud-Platform/k8s-workload-operator/api/v1alpha1"
	"github.com/KETI-Cloud-Platform/k8s-workload-operator/test/utils"
)

var _ = Describe("WorkloadOptimizer E2E Tests", Ordered, func() {
	var (
		ctx         context.Context
		k8sClient   client.Client
		testNs      string
		workload    *kcloudv1alpha1.WorkloadOptimizer
		costPolicy  *kcloudv1alpha1.CostPolicy
		powerPolicy *kcloudv1alpha1.PowerPolicy
	)

	BeforeAll(func() {
		ctx = context.Background()
		testNs = "kcloud-e2e-test"

		// Create test namespace
		cmd := exec.Command("kubectl", "create", "ns", testNs)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())

		// Create test nodes
		createTestNodes()
	})

	AfterAll(func() {
		// Clean up test namespace
		cmd := exec.Command("kubectl", "delete", "ns", testNs, "--ignore-not-found=true")
		_, _ = utils.Run(cmd)
	})

	Context("WorkloadOptimizer CRD Operations", func() {
		It("should create and reconcile WorkloadOptimizer successfully", func() {
			By("creating a WorkloadOptimizer resource")
			workload = &kcloudv1alpha1.WorkloadOptimizer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload",
					Namespace: testNs,
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
					AutoScaling: &kcloudv1alpha1.AutoScalingSpec{
						Enabled:      true,
						MinReplicas:  1,
						MaxReplicas:  5,
						TargetCPU:    70,
						TargetMemory: 80,
					},
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(workload)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for WorkloadOptimizer to be reconciled")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "workloadoptimizer", "test-workload", "-n", testNs, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				return err == nil && output != ""
			}, 2*time.Minute, 10*time.Second).Should(BeTrue())

			By("verifying WorkloadOptimizer status")
			cmd = exec.Command("kubectl", "get", "workloadoptimizer", "test-workload", "-n", testNs, "-o", "jsonpath={.status.phase}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(BeElementOf("Pending", "Optimizing", "Optimized"))
		})

		It("should update WorkloadOptimizer status with optimization results", func() {
			By("waiting for optimization score to be set")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "workloadoptimizer", "test-workload", "-n", testNs, "-o", "jsonpath={.status.optimizationScore}")
				output, err := utils.Run(cmd)
				return err == nil && output != ""
			}, 3*time.Minute, 15*time.Second).Should(BeTrue())

			By("verifying optimization metrics")
			cmd := exec.Command("kubectl", "get", "workloadoptimizer", "test-workload", "-n", testNs, "-o", "jsonpath={.status.optimizationScore}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("0.")) // Should be a decimal between 0 and 1
		})

		It("should handle WorkloadOptimizer deletion", func() {
			By("deleting the WorkloadOptimizer")
			cmd := exec.Command("kubectl", "delete", "workloadoptimizer", "test-workload", "-n", testNs)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("verifying WorkloadOptimizer is deleted")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "workloadoptimizer", "test-workload", "-n", testNs)
				_, err := utils.Run(cmd)
				return err != nil // Should fail when resource doesn't exist
			}, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	Context("CostPolicy CRD Operations", func() {
		It("should create and manage CostPolicy successfully", func() {
			By("creating a CostPolicy resource")
			costPolicy = &kcloudv1alpha1.CostPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cost-policy",
					Namespace: testNs,
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
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"cost-policy": "enabled",
						},
					},
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(costPolicy)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for CostPolicy to be processed")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "costpolicy", "test-cost-policy", "-n", testNs, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				return err == nil && output != ""
			}, 1*time.Minute, 5*time.Second).Should(BeTrue())

			By("verifying CostPolicy status")
			cmd = exec.Command("kubectl", "get", "costpolicy", "test-cost-policy", "-n", testNs, "-o", "jsonpath={.status.phase}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(BeElementOf("Active", "Pending"))
		})
	})

	Context("PowerPolicy CRD Operations", func() {
		It("should create and manage PowerPolicy successfully", func() {
			By("creating a PowerPolicy resource")
			powerPolicy = &kcloudv1alpha1.PowerPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-power-policy",
					Namespace: testNs,
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
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"power-policy": "enabled",
						},
					},
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(powerPolicy)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for PowerPolicy to be processed")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "powerpolicy", "test-power-policy", "-n", testNs, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				return err == nil && output != ""
			}, 1*time.Minute, 5*time.Second).Should(BeTrue())

			By("verifying PowerPolicy status")
			cmd = exec.Command("kubectl", "get", "powerpolicy", "test-power-policy", "-n", testNs, "-o", "jsonpath={.status.phase}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(BeElementOf("Active", "Pending"))
		})
	})

	Context("Webhook Integration", func() {
		It("should validate WorkloadOptimizer through webhook", func() {
			By("creating an invalid WorkloadOptimizer")
			invalidWorkload := &kcloudv1alpha1.WorkloadOptimizer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-workload",
					Namespace: testNs,
				},
				Spec: kcloudv1alpha1.WorkloadOptimizerSpec{
					WorkloadType: "invalid-type", // Invalid workload type
					Priority:     15,             // Invalid priority (should be 1-10)
					ResourceRequirements: kcloudv1alpha1.ResourceRequirements{
						CPU:    resource.MustParse("0"), // Invalid CPU
						Memory: resource.MustParse("0"), // Invalid memory
						GPU:    20,                      // Invalid GPU count
						NPU:    20,                      // Invalid NPU count
					},
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(invalidWorkload)
			output, err := utils.Run(cmd)

			// Should fail due to validation
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("admission webhook"))
		})

		It("should mutate Pods based on WorkloadOptimizer policies", func() {
			By("creating a WorkloadOptimizer with specific policies")
			mutateWorkload := &kcloudv1alpha1.WorkloadOptimizer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mutate-workload",
					Namespace: testNs,
				},
				Spec: kcloudv1alpha1.WorkloadOptimizerSpec{
					WorkloadType: "training",
					Priority:     8,
					ResourceRequirements: kcloudv1alpha1.ResourceRequirements{
						CPU:    resource.MustParse("4"),
						Memory: resource.MustParse("8Gi"),
						GPU:    2,
						NPU:    0,
					},
					PlacementPolicy: &kcloudv1alpha1.PlacementPolicy{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "node-type",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"gpu-optimized"},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(mutateWorkload)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("creating a Pod that should be mutated")
			testPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: testNs,
					Labels: map[string]string{
						"workloadoptimizer.kcloud.io/name": "mutate-workload",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
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

			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(testPod)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("verifying Pod was mutated")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "pod", "test-pod", "-n", testNs, "-o", "jsonpath={.spec.nodeSelector}")
				output, err := utils.Run(cmd)
				return err == nil && output != ""
			}, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	Context("Metrics Integration", func() {
		It("should expose Prometheus metrics", func() {
			By("creating a WorkloadOptimizer to generate metrics")
			metricsWorkload := &kcloudv1alpha1.WorkloadOptimizer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "metrics-workload",
					Namespace: testNs,
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
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(metricsWorkload)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for metrics to be generated")
			time.Sleep(30 * time.Second)

			By("verifying metrics endpoint contains kcloud metrics")
			metricsOutput, err := getMetricsOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(metricsOutput).To(ContainSubstring("kcloud_workloadoptimizer"))
		})
	})

	Context("End-to-End Workflow", func() {
		It("should complete full optimization workflow", func() {
			By("creating a comprehensive WorkloadOptimizer")
			fullWorkload := &kcloudv1alpha1.WorkloadOptimizer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "full-workload",
					Namespace: testNs,
				},
				Spec: kcloudv1alpha1.WorkloadOptimizerSpec{
					WorkloadType: "training",
					Priority:     7,
					ResourceRequirements: kcloudv1alpha1.ResourceRequirements{
						CPU:    resource.MustParse("4"),
						Memory: resource.MustParse("8Gi"),
						GPU:    2,
						NPU:    0,
					},
					CostConstraints: kcloudv1alpha1.CostConstraints{
						MaxCostPerHour: 15.0,
						BudgetLimit:    2000.0,
					},
					PowerConstraints: kcloudv1alpha1.PowerConstraints{
						MaxPowerUsage: 800.0,
					},
					PlacementPolicy: &kcloudv1alpha1.PlacementPolicy{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "node-type",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"gpu-optimized"},
											},
										},
									},
								},
							},
						},
					},
					AutoScaling: &kcloudv1alpha1.AutoScalingSpec{
						Enabled:      true,
						MinReplicas:  2,
						MaxReplicas:  10,
						TargetCPU:    75,
						TargetMemory: 85,
					},
				},
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.YAMLToReader(fullWorkload)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("waiting for full reconciliation")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "workloadoptimizer", "full-workload", "-n", testNs, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				return err == nil && output == "Optimized"
			}, 5*time.Minute, 30*time.Second).Should(BeTrue())

			By("verifying all status fields are populated")
			cmd = exec.Command("kubectl", "get", "workloadoptimizer", "full-workload", "-n", testNs, "-o", "jsonpath={.status}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("optimizationScore"))
			Expect(output).To(ContainSubstring("currentCost"))
			Expect(output).To(ContainSubstring("currentPower"))
			Expect(output).To(ContainSubstring("assignedNode"))
		})
	})
})

// createTestNodes creates test nodes for E2E testing
func createTestNodes() {
	nodes := []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "e2e-cpu-node",
				Labels: map[string]string{
					"node-type": "cpu-optimized",
				},
			},
			Status: corev1.NodeStatus{
				Allocatable: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("8"),
					corev1.ResourceMemory: resource.MustParse("16Gi"),
					corev1.ResourceGPU:    resource.MustParse("0"),
				},
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "e2e-gpu-node",
				Labels: map[string]string{
					"node-type": "gpu-optimized",
				},
			},
			Status: corev1.NodeStatus{
				Allocatable: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("16"),
					corev1.ResourceMemory: resource.MustParse("32Gi"),
					corev1.ResourceGPU:    resource.MustParse("4"),
				},
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
	}

	for _, node := range nodes {
		cmd := exec.Command("kubectl", "apply", "-f", "-")
		cmd.Stdin = utils.YAMLToReader(node)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())
	}
}
