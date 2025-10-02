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

package optimizer

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kcloudv1alpha1 "github.com/KETI-Cloud-Platform/k8s-workload-operator/api/v1alpha1"
)

func TestOptimizerEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Optimizer Engine Suite")
}

var _ = Describe("Optimizer Engine", func() {
	var (
		ctx      context.Context
		engine   *Engine
		workload *kcloudv1alpha1.WorkloadOptimizer
	)

	BeforeEach(func() {
		ctx = context.Background()
		engine = NewEngine()

		workload = &kcloudv1alpha1.WorkloadOptimizer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workload",
				Namespace: "test-namespace",
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
	})

	Context("When optimizing workload", func() {
		It("should optimize workload successfully", func() {
			// Create test workload state
			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			// Optimize
			result, err := engine.Optimize(ctx, state)

			// Verify result
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Score).To(BeNumerically(">=", 0))
			Expect(result.Score).To(BeNumerically("<=", 1))
			Expect(result.EstimatedCost).To(BeNumerically(">=", 0))
			Expect(result.EstimatedPower).To(BeNumerically(">=", 0))
		})

		It("should respect cost constraints", func() {
			// Create test workload state
			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			// Optimize
			result, err := engine.Optimize(ctx, state)

			// Verify cost constraints
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.EstimatedCost).To(BeNumerically("<=", 10.0))
		})

		It("should respect power constraints", func() {
			// Create test workload state
			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			// Optimize
			result, err := engine.Optimize(ctx, state)

			// Verify power constraints
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.EstimatedPower).To(BeNumerically("<=", 500.0))
		})

		It("should handle different workload types", func() {
			workloadTypes := []string{"training", "inference", "batch", "streaming"}

			for _, workloadType := range workloadTypes {
				workload.Spec.WorkloadType = workloadType
				state := &WorkloadState{
					Workload: workload,
					Pods:     []corev1.Pod{},
					Nodes:    []corev1.Node{},
				}

				result, err := engine.Optimize(ctx, state)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Score).To(BeNumerically(">=", 0))
				Expect(result.Score).To(BeNumerically("<=", 1))
			}
		})

		It("should handle different priorities", func() {
			priorities := []int{1, 5, 10}

			for _, priority := range priorities {
				workload.Spec.Priority = priority
				state := &WorkloadState{
					Workload: workload,
					Pods:     []corev1.Pod{},
					Nodes:    []corev1.Node{},
				}

				result, err := engine.Optimize(ctx, state)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Score).To(BeNumerically(">=", 0))
				Expect(result.Score).To(BeNumerically("<=", 1))
			}
		})

		It("should handle empty workload state", func() {
			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})

		It("should handle nil workload", func() {
			state := &WorkloadState{
				Workload: nil,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("When calculating optimization score", func() {
		It("should calculate score based on resource utilization", func() {
			// Create test workload state with pods
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
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

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{*pod},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Score).To(BeNumerically(">=", 0))
			Expect(result.Score).To(BeNumerically("<=", 1))
		})

		It("should calculate score based on node availability", func() {
			// Create test node
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("8Gi"),
						corev1.ResourceGPU:    resource.MustParse("2"),
					},
				},
			}

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{*node},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Score).To(BeNumerically(">=", 0))
			Expect(result.Score).To(BeNumerically("<=", 1))
		})
	})

	Context("When handling constraints", func() {
		It("should handle cost constraint violations", func() {
			// Set very low cost constraint
			workload.Spec.CostConstraints.MaxCostPerHour = 0.1

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// Score should be lower due to constraint violation
			Expect(result.Score).To(BeNumerically("<", 1.0))
		})

		It("should handle power constraint violations", func() {
			// Set very low power constraint
			workload.Spec.PowerConstraints.MaxPowerUsage = 10.0

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// Score should be lower due to constraint violation
			Expect(result.Score).To(BeNumerically("<", 1.0))
		})

		It("should handle both cost and power constraints", func() {
			// Set both constraints
			workload.Spec.CostConstraints.MaxCostPerHour = 5.0
			workload.Spec.PowerConstraints.MaxPowerUsage = 250.0

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.EstimatedCost).To(BeNumerically("<=", 5.0))
			Expect(result.EstimatedPower).To(BeNumerically("<=", 250.0))
		})
	})

	Context("When handling auto-scaling", func() {
		It("should handle auto-scaling enabled", func() {
			workload.Spec.AutoScaling = &kcloudv1alpha1.AutoScalingSpec{
				Enabled:      true,
				MinReplicas:  1,
				MaxReplicas:  10,
				TargetCPU:    70,
				TargetMemory: 80,
			}

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.RecommendedReplicas).To(BeNumerically(">=", 1))
			Expect(result.RecommendedReplicas).To(BeNumerically("<=", 10))
		})

		It("should handle auto-scaling disabled", func() {
			workload.Spec.AutoScaling = &kcloudv1alpha1.AutoScalingSpec{
				Enabled:      false,
				MinReplicas:  1,
				MaxReplicas:  10,
				TargetCPU:    70,
				TargetMemory: 80,
			}

			state := &WorkloadState{
				Workload: workload,
				Pods:     []corev1.Pod{},
				Nodes:    []corev1.Node{},
			}

			result, err := engine.Optimize(ctx, state)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.RecommendedReplicas).To(Equal(1))
		})
	})
})
