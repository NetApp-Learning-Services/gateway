package v1beta3

type NvmeSubSpec struct {
	// Provides required NVMe enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

	// Provides optional NVMe LIFs
	// +kubebuilder:validation:Optional
	Lifs []LIF `json:"interfaces,omitempty"`
}
