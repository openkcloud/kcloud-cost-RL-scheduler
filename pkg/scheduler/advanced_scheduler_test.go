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

func TestAdvancedScheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Advanced Scheduler Suite")
}

var _ = Describe("Advanced Scheduler", func() {
	var (
		advancedScheduler *AdvancedScheduler
		workload          *kcloudv1alpha1.WorkloadOptimizer
		nodes             []corev1.Node
	)

	BeforeEach(func() {
		advancedScheduler = NewAdvancedScheduler()

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
			},
		}

		// Create test nodes with different characteristics
		nodes = []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"node-type":    "cpu-optimized",
						"cost-tier":    "low",
						"power-tier":   "medium",
						"availability": "high",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("8Gi"),
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
					Name: "node-2",
					Labels: map[string]string{
						"node-type":    "gpu-optimized",
						"cost-tier":    "high",
						"power-tier":   "high",
						"availability": "medium",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("8"),
						corev1.ResourceMemory: resource.MustParse("16Gi"),
						corev1.ResourceGPU:    resource.MustParse("2"),
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
						"node-type":    "npu-optimized",
						"cost-tier":    "medium",
						"power-tier":   "low",
						"availability": "high",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("6"),
						corev1.ResourceMemory: resource.MustParse("12Gi"),
						corev1.ResourceNPU:    resource.MustParse("1"),
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

	Context("When using round-robin scheduling", func() {
		It("should distribute workloads evenly", func() {
			algorithm := "round-robin"

			// Schedule multiple workloads
			results := make([]*SchedulingResult, 3)
			for i := 0; i < 3; i++ {
				result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				results[i] = result
			}

			// Should distribute across different nodes
			nodeNames := make(map[string]int)
			for _, result := range results {
				nodeNames[result.SelectedNode.Name]++
			}

			// Should have used multiple nodes
			Expect(len(nodeNames)).To(BeNumerically(">", 1))
		})
	})

	Context("When using least-loaded scheduling", func() {
		It("should select least loaded node", func() {
			algorithm := "least-loaded"

			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			Expect(result.Score).To(BeNumerically(">=", 0))
			Expect(result.Score).To(BeNumerically("<=", 1))
		})
	})

	Context("When using cost-optimized scheduling", func() {
		It("should prefer low-cost nodes", func() {
			algorithm := "cost-optimized"

			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())

			// Should prefer node-1 (low cost) over node-2 (high cost)
			Expect(result.SelectedNode.Name).To(Equal("node-1"))
		})
	})

	Context("When using power-optimized scheduling", func() {
		It("should prefer low-power nodes", func() {
			algorithm := "power-optimized"

			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())

			// Should prefer node-3 (low power) over node-2 (high power)
			Expect(result.SelectedNode.Name).To(Equal("node-3"))
		})
	})

	Context("When using balanced scheduling", func() {
		It("should balance multiple factors", func() {
			algorithm := "balanced"

			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode).NotTo(BeNil())
			Expect(result.Score).To(BeNumerically(">=", 0))
			Expect(result.Score).To(BeNumerically("<=", 1))
		})
	})

	Context("When using priority-based scheduling", func() {
		It("should consider workload priority", func() {
			algorithm := "priority-based"

			// High priority workload
			workload.Spec.Priority = 10
			result1, err1 := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			// Low priority workload
			workload.Spec.Priority = 1
			result2, err2 := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			Expect(result1).NotTo(BeNil())
			Expect(result2).NotTo(BeNil())

			// High priority should get better score
			Expect(result1.Score).To(BeNumerically(">=", result2.Score))
		})
	})

	Context("When handling resource reservations", func() {
		It("should create resource reservation", func() {
			reservation := &ResourceReservation{
				WorkloadName: workload.Name,
				Namespace:    workload.Namespace,
				NodeName:     "node-1",
				Resources:    workload.Spec.ResourceRequirements,
				Priority:     workload.Spec.Priority,
			}

			err := advancedScheduler.CreateReservation(reservation)
			Expect(err).NotTo(HaveOccurred())

			// Verify reservation exists
			exists := advancedScheduler.HasReservation(workload.Name, workload.Namespace)
			Expect(exists).To(BeTrue())
		})

		It("should delete resource reservation", func() {
			reservation := &ResourceReservation{
				WorkloadName: workload.Name,
				Namespace:    workload.Namespace,
				NodeName:     "node-1",
				Resources:    workload.Spec.ResourceRequirements,
				Priority:     workload.Spec.Priority,
			}

			// Create reservation
			err := advancedScheduler.CreateReservation(reservation)
			Expect(err).NotTo(HaveOccurred())

			// Delete reservation
			err = advancedScheduler.DeleteReservation(workload.Name, workload.Namespace)
			Expect(err).NotTo(HaveOccurred())

			// Verify reservation is gone
			exists := advancedScheduler.HasReservation(workload.Name, workload.Namespace)
			Expect(exists).To(BeFalse())
		})

		It("should get reservations for node", func() {
			reservation := &ResourceReservation{
				WorkloadName: workload.Name,
				Namespace:    workload.Namespace,
				NodeName:     "node-1",
				Resources:    workload.Spec.ResourceRequirements,
				Priority:     workload.Spec.Priority,
			}

			err := advancedScheduler.CreateReservation(reservation)
			Expect(err).NotTo(HaveOccurred())

			reservations := advancedScheduler.GetReservationsForNode("node-1")
			Expect(reservations).To(HaveLen(1))
			Expect(reservations[0].WorkloadName).To(Equal(workload.Name))
		})
	})

	Context("When managing scheduling history", func() {
		It("should record scheduling event", func() {
			event := &SchedulingEvent{
				WorkloadName: workload.Name,
				Namespace:    workload.Namespace,
				Algorithm:    "round-robin",
				SelectedNode: "node-1",
				Score:        0.8,
				Timestamp:    metav1.Now(),
			}

			err := advancedScheduler.RecordSchedulingEvent(event)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get scheduling history", func() {
			event := &SchedulingEvent{
				WorkloadName: workload.Name,
				Namespace:    workload.Namespace,
				Algorithm:    "round-robin",
				SelectedNode: "node-1",
				Score:        0.8,
				Timestamp:    metav1.Now(),
			}

			err := advancedScheduler.RecordSchedulingEvent(event)
			Expect(err).NotTo(HaveOccurred())

			history := advancedScheduler.GetSchedulingHistory(workload.Name, workload.Namespace)
			Expect(history).To(HaveLen(1))
			Expect(history[0].WorkloadName).To(Equal(workload.Name))
		})

		It("should get algorithm statistics", func() {
			// Record multiple events with different algorithms
			algorithms := []string{"round-robin", "least-loaded", "cost-optimized"}
			for _, algorithm := range algorithms {
				event := &SchedulingEvent{
					WorkloadName: workload.Name,
					Namespace:    workload.Namespace,
					Algorithm:    algorithm,
					SelectedNode: "node-1",
					Score:        0.8,
					Timestamp:    metav1.Now(),
				}
				err := advancedScheduler.RecordSchedulingEvent(event)
				Expect(err).NotTo(HaveOccurred())
			}

			stats := advancedScheduler.GetAlgorithmStatistics()
			Expect(stats).NotTo(BeNil())
			Expect(len(stats)).To(Equal(3))
		})
	})

	Context("When checking advanced constraints", func() {
		It("should check node requirements", func() {
			node := nodes[1] // GPU node
			meetsRequirements := advancedScheduler.nodeMeetsRequirements(node, workload.Spec.ResourceRequirements)

			Expect(meetsRequirements).To(BeTrue())
		})

		It("should reject nodes without required resources", func() {
			node := nodes[0] // CPU-only node
			meetsRequirements := advancedScheduler.nodeMeetsRequirements(node, workload.Spec.ResourceRequirements)

			Expect(meetsRequirements).To(BeFalse())
		})

		It("should check node affinity", func() {
			// Add node affinity
			workload.Spec.PlacementPolicy = &kcloudv1alpha1.PlacementPolicy{
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
			}

			node1 := nodes[0] // CPU node
			node2 := nodes[1] // GPU node

			meetsAffinity1 := advancedScheduler.nodeMeetsAffinity(node1, workload.Spec.PlacementPolicy)
			meetsAffinity2 := advancedScheduler.nodeMeetsAffinity(node2, workload.Spec.PlacementPolicy)

			Expect(meetsAffinity1).To(BeFalse())
			Expect(meetsAffinity2).To(BeTrue())
		})

		It("should check node anti-affinity", func() {
			// Add node anti-affinity
			workload.Spec.PlacementPolicy = &kcloudv1alpha1.PlacementPolicy{
				NodeAntiAffinity: &corev1.NodeAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "workload-type",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"training"},
									},
								},
							},
							TopologyKey: "node-type",
						},
					},
				},
			}

			node := nodes[1] // GPU node
			meetsAntiAffinity := advancedScheduler.nodeMeetsAntiAffinity(node, workload.Spec.PlacementPolicy)

			Expect(meetsAntiAffinity).To(BeTrue())
		})
	})

	Context("When handling invalid inputs", func() {
		It("should handle invalid algorithm", func() {
			algorithm := "invalid-algorithm"

			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle empty node list", func() {
			algorithm := "round-robin"

			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, []corev1.Node{}, algorithm)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle nil workload", func() {
			algorithm := "round-robin"

			result, err := advancedScheduler.ScheduleWithAlgorithm(nil, nodes, algorithm)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("When handling node conditions", func() {
		It("should avoid nodes with disk pressure", func() {
			// Add disk pressure to node-1
			nodes[0].Status.Conditions = append(nodes[0].Status.Conditions, corev1.NodeCondition{
				Type:   corev1.NodeDiskPressure,
				Status: corev1.ConditionTrue,
			})

			algorithm := "least-loaded"
			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode.Name).NotTo(Equal("node-1"))
		})

		It("should avoid nodes with memory pressure", func() {
			// Add memory pressure to node-2
			nodes[1].Status.Conditions = append(nodes[1].Status.Conditions, corev1.NodeCondition{
				Type:   corev1.NodeMemoryPressure,
				Status: corev1.ConditionTrue,
			})

			algorithm := "least-loaded"
			result, err := advancedScheduler.ScheduleWithAlgorithm(workload, nodes, algorithm)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.SelectedNode.Name).NotTo(Equal("node-2"))
		})
	})
})
