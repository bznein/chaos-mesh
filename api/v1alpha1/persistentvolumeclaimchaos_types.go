package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// PersistentVolumeClaimChaos is the Schema for the persistentvolumeclaimchaos API
type PersistentVolumeClaimChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PersistentVolumeClaimChaosSpec   `json:"spec"`
	Status PersistentVolumeClaimChaosStatus `json:"status,omitempty"`
}

// PersistentVolumeClaimChaosSpec is the content of the specification for a PersistentVolumeClaimChaos
type PersistentVolumeClaimChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`
	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

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

	// Remove finalizers tell the chaos whether to patch the PV deleted and remove its finalizers
	// +optional
	RemoveFinalizers bool `json:"remove_finalizers"`
}

// PersistentVolumeClaimChaosStatus represents the status of a PersistentVolumeClaimChaos
type PersistentVolumeClaimChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (in *PersistentVolumeClaimChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

func (in *PersistentVolumeClaimChaosSpec) GetMode() PodMode {
	return in.Mode
}

func (in *PersistentVolumeClaimChaosSpec) GetValue() string {
	return in.Value
}
