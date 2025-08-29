package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkloadOptimizerSpec defines the desired state of WorkloadOptimizer
type WorkloadOptimizerSpec struct {
	// WorkloadType specifies the type of workload (training, inference, serving)
	WorkloadType string `json:"workloadType"`
	
	// Priority of the workload (0-100)
	Priority int32 `json:"priority,omitempty"`
	
	// Resources required by the workload
	Resources ResourceRequirements `json:"resources"`
	
	// CostConstraints for the workload
	CostConstraints CostConstraints `json:"costConstraints,omitempty"`
	
	// PowerConstraints for the workload
	PowerConstraints PowerConstraints `json:"powerConstraints,omitempty"`
	
	// PlacementPolicy for the workload
	PlacementPolicy PlacementPolicy `json:"placementPolicy,omitempty"`
	
	// AutoScaling configuration
	AutoScaling *AutoScalingSpec `json:"autoScaling,omitempty"`
}

// ResourceRequirements defines the compute resources required
type ResourceRequirements struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	GPU    int32  `json:"gpu,omitempty"`
	NPU    int32  `json:"npu,omitempty"`
}

// CostConstraints defines cost-related constraints
type CostConstraints struct {
	MaxCostPerHour  float64 `json:"maxCostPerHour,omitempty"`
	PreferSpot      bool    `json:"preferSpot,omitempty"`
	BudgetLimit     float64 `json:"budgetLimit,omitempty"`
}

// PowerConstraints defines power-related constraints
type PowerConstraints struct {
	MaxPowerUsage float64 `json:"maxPowerUsage,omitempty"`
	PreferGreen   bool    `json:"preferGreen,omitempty"`
}

// PlacementPolicy defines placement preferences
type PlacementPolicy struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Affinity     []AffinityRule    `json:"affinity,omitempty"`
	Tolerations  []Toleration      `json:"tolerations,omitempty"`
}

// AffinityRule defines affinity rules
type AffinityRule struct {
	Type   string `json:"type"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Weight int32  `json:"weight,omitempty"`
}

// Toleration defines pod tolerations
type Toleration struct {
	Key      string `json:"key"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// AutoScalingSpec defines auto-scaling configuration
type AutoScalingSpec struct {
	MinReplicas int32 `json:"minReplicas"`
	MaxReplicas int32 `json:"maxReplicas"`
	Metrics     []AutoScalingMetric `json:"metrics"`
}

// AutoScalingMetric defines metrics for auto-scaling
type AutoScalingMetric struct {
	Type       string `json:"type"`
	Threshold  int32  `json:"threshold"`
}

// WorkloadOptimizerStatus defines the observed state of WorkloadOptimizer
type WorkloadOptimizerStatus struct {
	// Phase of the workload
	Phase string `json:"phase,omitempty"`
	
	// Current cost per hour
	CurrentCost float64 `json:"currentCost,omitempty"`
	
	// Current power usage
	CurrentPower float64 `json:"currentPower,omitempty"`
	
	// Assigned node
	AssignedNode string `json:"assignedNode,omitempty"`
	
	// Optimization score
	OptimizationScore float64 `json:"optimizationScore,omitempty"`
	
	// Last optimization time
	LastOptimized metav1.Time `json:"lastOptimized,omitempty"`
	
	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=wo
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.workloadType`
//+kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`
//+kubebuilder:printcolumn:name="Cost",type=number,JSONPath=`.status.currentCost`
//+kubebuilder:printcolumn:name="Power",type=number,JSONPath=`.status.currentPower`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// WorkloadOptimizer is the Schema for the workloadoptimizers API
type WorkloadOptimizer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkloadOptimizerSpec   `json:"spec,omitempty"`
	Status WorkloadOptimizerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkloadOptimizerList contains a list of WorkloadOptimizer
type WorkloadOptimizerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkloadOptimizer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkloadOptimizer{}, &WorkloadOptimizerList{})
}