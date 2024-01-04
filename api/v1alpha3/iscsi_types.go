package v1alpha3

type IscsiSubSpec struct {
	// Provides required iSCSI enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

	// Provides optional iSCSI LIFs
	// +kubebuilder:validation:Optional
	Lifs []LIF `json:"interfaces,omitempty"`

	// Provides optional alias
	// +kubebuilder:validation:Optional
	Alias string `json:"alias,omitempty"`
}
