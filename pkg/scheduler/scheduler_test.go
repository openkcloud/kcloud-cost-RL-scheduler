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

package scheduler

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kcloudv1alpha1 "github.com/KETI-Cloud-Platform/k8s-workload-operator/api/v1alpha1"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheduler Suite")
}

var _ = Describe("Scheduler", func() {
	var (
		scheduler *Scheduler
		workload  *kcloudv1alpha1.WorkloadOptimizer
		nodes     []corev1.Node
	)

	BeforeEach(func() {
		scheduler = NewScheduler()

		workload = &kcloudv1alpha1.WorkloadOptimizer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workload",
				Namespace: "test-namespace",
			},
			Spec: kcloudv1alpha1.WorkloadOptimizerSpec{
				WorkloadType: "training",
				Priority:     5,
				Resources: kcloudv1alpha1.ResourceRequirements{
					CPU:    "2",
					Memory: "4Gi",
					GPU:    1,
					NPU:    0,
				},
			},
		}

		// Create test nodes
		nodes = []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"node-type": "cpu-optimized",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("8Gi"),
						"nvidia.com/gpu":      resource.MustParse("0"),
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
					Name: "node-2",
					Labels: map[string]string{
						"node-type": "gpu-optimized",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("8"),
						corev1.ResourceMemory: resource.MustParse("16Gi"),
						"nvidia.com/gpu":      resource.MustParse("2"),
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
					Name: "node-3",
					Labels: map[string]string{
						"node-type": "npu-optimized",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("6"),
						corev1.ResourceMemory: resource.MustParse("12Gi"),
						"huawei.com/npu":      resource.MustParse("1"),
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
	})

	Context("When scheduling workload", func() {
		It("should schedule workload to appropriate node", func() {
			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			Expect(result.Score).To(BeNumerically(">=", 0))
			Expect(result.Score).To(BeNumerically("<=", 1))
		})

		It("should prefer GPU node for GPU workload", func() {
			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			// Should prefer node-2 (GPU node) for GPU workload
			Expect(result.SelectedNode.Name).To(Equal("node-2"))
		})

		It("should prefer NPU node for NPU workload", func() {
			// Modify workload to require NPU
			workload.Spec.ResourceRequirements.GPU = 0
			workload.Spec.ResourceRequirements.NPU = 1

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			// Should prefer node-3 (NPU node) for NPU workload
			Expect(result.SelectedNode.Name).To(Equal("node-3"))
		})

		It("should handle CPU-only workload", func() {
			// Modify workload to require only CPU
			workload.Spec.ResourceRequirements.GPU = 0
			workload.Spec.ResourceRequirements.NPU = 0

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			// Any node should work for CPU-only workload
			Expect(result.SelectedNode.Name).To(BeElementOf("node-1", "node-2", "node-3"))
		})

		It("should handle empty node list", func() {
			result, err := scheduler.ScheduleWorkload(ctx, workload, []corev1.Node{})

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle nil workload", func() {
			result, err := scheduler.ScheduleWorkload(ctx, nil, nodes)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle nodes without required resources", func() {
			// Create nodes without GPU
			cpuOnlyNodes := []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cpu-only-node",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("4"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
							"nvidia.com/gpu":      resource.MustParse("0"),
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

			result, err := scheduler.ScheduleWorkload(ctx, workload, cpuOnlyNodes)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("When evaluating node suitability", func() {
		It("should evaluate node resource availability", func() {
			node := nodes[1] // GPU node
			score, err := scheduler.evaluateNode(ctx, workload, node)

			Expect(score).To(BeNumerically(">=", 0))
			Expect(score).To(BeNumerically("<=", 1))
		})

		It("should give higher score to nodes with more resources", func() {
			score1 := scheduler.evaluateNode(workload, nodes[0]) // CPU node
			score2 := scheduler.evaluateNode(workload, nodes[1]) // GPU node

			Expect(score2).To(BeNumerically(">", score1))
		})

		It("should consider node labels and affinity", func() {
			// Add node affinity to workload
			workload.Spec.PlacementPolicy = &kcloudv1alpha1.PlacementPolicy{
				NodeSelector: map[string]string{
					"node-type": "gpu-optimized",
				},
			}

			score1 := scheduler.evaluateNode(workload, nodes[0]) // CPU node
			score2 := scheduler.evaluateNode(workload, nodes[1]) // GPU node

			Expect(score2).To(BeNumerically(">", score1))
		})
	})

	Context("When handling different workload types", func() {
		It("should schedule training workloads appropriately", func() {
			workload.Spec.WorkloadType = "training"
			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})

		It("should schedule inference workloads appropriately", func() {
			workload.Spec.WorkloadType = "inference"
			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})

		It("should schedule batch workloads appropriately", func() {
			workload.Spec.WorkloadType = "batch"
			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})

		It("should schedule streaming workloads appropriately", func() {
			workload.Spec.WorkloadType = "streaming"
			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})
	})

	Context("When handling different priorities", func() {
		It("should consider workload priority", func() {
			workload.Spec.Priority = 10 // High priority
			result1, err1 := scheduler.ScheduleWorkload(ctx, workload, nodes)

			workload.Spec.Priority = 1 // Low priority
			result2, err2 := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			Expect(result1).NotTo(BeNil())
			Expect(result2).NotTo(BeNil())

			// High priority workload should get better node
			Expect(result1.Score).To(BeNumerically(">=", result2.Score))
		})
	})

	Context("When handling constraints", func() {
		It("should respect cost constraints", func() {
			workload.Spec.CostConstraints = kcloudv1alpha1.CostConstraints{
				MaxCostPerHour: 5.0,
				BudgetLimit:    100.0,
			}

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})

		It("should respect power constraints", func() {
			workload.Spec.PowerConstraints = kcloudv1alpha1.PowerConstraints{
				MaxPowerUsage: 200.0,
			}

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})
	})

	Context("When handling node conditions", func() {
		It("should avoid nodes that are not ready", func() {
			// Make node-1 not ready
			nodes[0].Status.Conditions = []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			}

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			// Should not select node-1
			Expect(result.SelectedNode.Name).NotTo(Equal("node-1"))
		})

		It("should handle nodes with disk pressure", func() {
			// Add disk pressure to node-1
			nodes[0].Status.Conditions = append(nodes[0].Status.Conditions, corev1.NodeCondition{
				Type:   corev1.NodeDiskPressure,
				Status: corev1.ConditionTrue,
			})

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			// Should prefer other nodes
			Expect(result.SelectedNode.Name).NotTo(Equal("node-1"))
		})
	})

	Context("When handling resource requirements", func() {
		It("should handle large resource requirements", func() {
			workload.Spec.ResourceRequirements = kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("16"),
				Memory: resource.MustParse("32Gi"),
				GPU:    4,
				NPU:    2,
			}

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle minimal resource requirements", func() {
			workload.Spec.ResourceRequirements = kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("0.1"),
				Memory: resource.MustParse("100Mi"),
				GPU:    0,
				NPU:    0,
			}

			result, err := scheduler.ScheduleWorkload(ctx, workload, nodes)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
		})
	})
})
