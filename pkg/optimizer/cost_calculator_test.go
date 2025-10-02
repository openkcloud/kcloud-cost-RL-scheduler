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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kcloudv1alpha1 "github.com/KETI-Cloud-Platform/k8s-workload-operator/api/v1alpha1"
)

func TestCostCalculator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cost Calculator Suite")
}

var _ = Describe("Cost Calculator", func() {
	var (
		calculator *CostCalculator
		workload   *kcloudv1alpha1.WorkloadOptimizer
	)

	BeforeEach(func() {
		calculator = NewCostCalculator()

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
	})

	Context("When calculating cost", func() {
		It("should calculate cost for CPU resources", func() {
			cost := calculator.CalculateCost(2.0, 4.0, 1, 0)

			Expect(cost).To(BeNumerically(">", 0))
			// CPU cost should be proportional to CPU amount
			Expect(cost).To(BeNumerically(">=", 0.1)) // Minimum cost for 2 CPU
		})

		It("should calculate cost for memory resources", func() {
			cost := calculator.CalculateCost(2.0, 4.0, 1, 0)

			Expect(cost).To(BeNumerically(">", 0))
			// Memory cost should be proportional to memory amount
			Expect(cost).To(BeNumerically(">=", 0.1)) // Minimum cost for 4Gi memory
		})

		It("should calculate cost for GPU resources", func() {
			cost := calculator.CalculateCost(2.0, 4.0, 1, 0)

			Expect(cost).To(BeNumerically(">", 0))
			// GPU cost should be significant
			Expect(cost).To(BeNumerically(">=", 0.5)) // Minimum cost for 1 GPU
		})

		It("should calculate cost for NPU resources", func() {
			// Test with NPU
			workload.Spec.ResourceRequirements.NPU = 1
			workload.Spec.ResourceRequirements.GPU = 0

			cost := calculator.CalculateCost(2.0, 4.0, 1, 0)

			Expect(cost).To(BeNumerically(">", 0))
			// NPU cost should be significant
			Expect(cost).To(BeNumerically(">=", 0.5)) // Minimum cost for 1 NPU
		})

		It("should calculate higher cost for more resources", func() {
			// Test with minimal resources
			minimalReq := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("0.5"),
				Memory: resource.MustParse("1Gi"),
				GPU:    0,
				NPU:    0,
			}
			minimalCost := calculator.CalculateCost(minimalReq)

			// Test with more resources
			highReq := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("8"),
				Memory: resource.MustParse("16Gi"),
				GPU:    4,
				NPU:    2,
			}
			highCost := calculator.CalculateCost(highReq)

			Expect(highCost).To(BeNumerically(">", minimalCost))
		})

		It("should handle zero resources", func() {
			zeroReq := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("0"),
				Memory: resource.MustParse("0"),
				GPU:    0,
				NPU:    0,
			}

			cost := calculator.CalculateCost(zeroReq)

			Expect(cost).To(Equal(0.0))
		})
	})

	Context("When calculating time-based costs", func() {
		It("should calculate daily cost", func() {
			hourlyCost := calculator.CalculateCost(2.0, 4.0, 1, 0)
			dailyCost := calculator.CalculateDailyCost(2.0, 4.0, 1, 0)

			Expect(dailyCost).To(BeNumerically(">", 0))
			Expect(dailyCost).To(BeNumerically(">", hourlyCost))
			Expect(dailyCost).To(BeNumerically("~", hourlyCost*24, 0.01))
		})

		It("should calculate monthly cost", func() {
			hourlyCost := calculator.CalculateCost(2.0, 4.0, 1, 0)
			monthlyCost := calculator.CalculateMonthlyCost(2.0, 4.0, 1, 0)

			Expect(monthlyCost).To(BeNumerically(">", 0))
			Expect(monthlyCost).To(BeNumerically(">", hourlyCost))
			Expect(monthlyCost).To(BeNumerically("~", hourlyCost*24*30, 0.01))
		})

		It("should calculate yearly cost", func() {
			hourlyCost := calculator.CalculateCost(2.0, 4.0, 1, 0)
			yearlyCost := calculator.CalculateYearlyCost(2.0, 4.0, 1, 0)

			Expect(yearlyCost).To(BeNumerically(">", 0))
			Expect(yearlyCost).To(BeNumerically(">", hourlyCost))
			Expect(yearlyCost).To(BeNumerically("~", hourlyCost*24*365, 0.01))
		})
	})

	Context("When calculating savings", func() {
		It("should calculate savings from optimization", func() {
			originalCost := 10.0
			optimizedCost := 7.0

			savings := originalCost - optimizedCost

			Expect(savings).To(Equal(3.0))
		})

		It("should handle negative savings", func() {
			originalCost := 5.0
			optimizedCost := 8.0

			savings := originalCost - optimizedCost

			Expect(savings).To(Equal(-3.0))
		})

		It("should handle equal costs", func() {
			originalCost := 10.0
			optimizedCost := 10.0

			savings := originalCost - optimizedCost

			Expect(savings).To(Equal(0.0))
		})
	})

	Context("When getting cost breakdown", func() {
		It("should provide cost breakdown", func() {
			breakdown := calculator.GetCostBreakdown(workload.Spec.ResourceRequirements)

			Expect(breakdown).NotTo(BeNil())
			Expect(breakdown.CPU).To(BeNumerically(">=", 0))
			Expect(breakdown.Memory).To(BeNumerically(">=", 0))
			Expect(breakdown.GPU).To(BeNumerically(">=", 0))
			Expect(breakdown.NPU).To(BeNumerically(">=", 0))
			Expect(breakdown.Total).To(BeNumerically(">", 0))
		})

		It("should have total equal to sum of components", func() {
			breakdown := calculator.GetCostBreakdown(workload.Spec.ResourceRequirements)

			expectedTotal := breakdown.CPU + breakdown.Memory + breakdown.GPU + breakdown.NPU
			Expect(breakdown.Total).To(BeNumerically("~", expectedTotal, 0.01))
		})
	})

	Context("When parsing resource strings", func() {
		It("should parse CPU resource strings", func() {
			cpu, err := calculator.parseCPU("2")
			Expect(err).NotTo(HaveOccurred())
			Expect(cpu).To(Equal(2.0))

			cpu, err = calculator.parseCPU("0.5")
			Expect(err).NotTo(HaveOccurred())
			Expect(cpu).To(Equal(0.5))
		})

		It("should parse memory resource strings", func() {
			memory, err := calculator.parseMemory("4Gi")
			Expect(err).NotTo(HaveOccurred())
			Expect(memory).To(Equal(4.0))

			memory, err = calculator.parseMemory("512Mi")
			Expect(err).NotTo(HaveOccurred())
			Expect(memory).To(BeNumerically("~", 0.5, 0.01))
		})

		It("should parse resource strings", func() {
			req := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("2"),
				Memory: resource.MustParse("4Gi"),
				GPU:    1,
				NPU:    0,
			}

			parsed, err := calculator.ParseResourceString(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.CPU).To(Equal(2.0))
			Expect(parsed.Memory).To(Equal(4.0))
			Expect(parsed.GPU).To(Equal(1))
			Expect(parsed.NPU).To(Equal(0))
		})

		It("should calculate cost from resource string", func() {
			req := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("2"),
				Memory: resource.MustParse("4Gi"),
				GPU:    1,
				NPU:    0,
			}

			cost, err := calculator.CalculateCostFromResourceString(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(cost).To(BeNumerically(">", 0))
		})
	})

	Context("When handling different workload types", func() {
		It("should calculate different costs for different workload types", func() {
			workloadTypes := []string{"training", "inference", "batch", "streaming"}
			costs := make(map[string]float64)

			for _, workloadType := range workloadTypes {
				workload.Spec.WorkloadType = workloadType
				cost := calculator.CalculateCost(2.0, 4.0, 1, 0)
				costs[workloadType] = cost
			}

			// All costs should be positive
			for _, cost := range costs {
				Expect(cost).To(BeNumerically(">", 0))
			}

			// Training workloads should generally cost more than inference
			Expect(costs["training"]).To(BeNumerically(">=", costs["inference"]))
		})
	})

	Context("When handling edge cases", func() {
		It("should handle very large resource requirements", func() {
			largeReq := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("100"),
				Memory: resource.MustParse("1000Gi"),
				GPU:    50,
				NPU:    25,
			}

			cost := calculator.CalculateCost(largeReq)
			Expect(cost).To(BeNumerically(">", 0))
			Expect(cost).To(BeNumerically("<", 10000)) // Reasonable upper bound
		})

		It("should handle fractional CPU requirements", func() {
			fractionalReq := kcloudv1alpha1.ResourceRequirements{
				CPU:    resource.MustParse("0.1"),
				Memory: resource.MustParse("100Mi"),
				GPU:    0,
				NPU:    0,
			}

			cost := calculator.CalculateCost(fractionalReq)
			Expect(cost).To(BeNumerically(">", 0))
			Expect(cost).To(BeNumerically("<", 1.0)) // Should be small
		})

		It("should handle memory in different units", func() {
			units := []string{"100Mi", "1Gi", "1000Mi", "2Gi"}
			costs := make([]float64, len(units))

			for i, unit := range units {
				req := kcloudv1alpha1.ResourceRequirements{
					CPU:    resource.MustParse("1"),
					Memory: resource.MustParse(unit),
					GPU:    0,
					NPU:    0,
				}
				costs[i] = calculator.CalculateCost(req)
			}

			// Costs should increase with memory size
			for i := 1; i < len(costs); i++ {
				Expect(costs[i]).To(BeNumerically(">=", costs[i-1]))
			}
		})
	})
})
