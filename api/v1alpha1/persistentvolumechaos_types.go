package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// PersistentVolumeChaos is the Schema for the persistentvolumechaos API
type PersistentVolumeChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PersistentVolumeChaosSpec   `json:"spec"`
	Status PersistentVolumeChaosStatus `json:"status,omitempty"`
}

// PersistentVolumeChaosSpec is the content of the specification for a PersistentVolumeChaos
type PersistentVolumeChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `fixed` / `fixed-percent` / `random-max-percent`.
	// If `fixed`, provide an integer value representign the number of volumes on which to act
	// If `fixed-percent`, provide a number from 0-100 to specify the percent of volumes on which to act
	// If `random-max-percent`,  provide a number from 0-100 to specify the max percent of volumes on which to act
	// +optional
	Value string `json:"value"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Remove finalizers tell the chaos whether to patch the PV deleted and remove its finalizers
	// +optional
	RemoveFinalizers bool `json:"remove_finalizers"`
}

// PersistentVolumeChaosStatus represents the status of a PersistentVolumeChaos
type PersistentVolumeChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (in *PersistentVolumeChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

func (in *PersistentVolumeChaosSpec) GetMode() PodMode {
	return in.Mode
}

func (in *PersistentVolumeChaosSpec) GetValue() string {
	return in.Value
}
