package v1alpha1

type NfsSubSpec struct {

	// Provides required NFS enablement
	// +kubebuilder:validation:Required
	NfsEnabled bool `json:"enabled"`

	// Provides optional NFS v3 enablement
	// +kubebuilder:validation:Optional
	Nfsv3 bool `json:"v3,omitempty"`

	// Provides optional NFS v4 enablement
	// +kubebuilder:validation:Optional
	Nfsv4 bool `json:"v4,omitempty"`

	// Provides optional NFS v4.1 enablement
	// +kubebuilder:validation:Optional
	Nfsv41 bool `json:"v41,omitempty"`

	// Provides optional NFS LIFs
	// +kubebuilder:validation:Optional
	NfsLifs []LIF `json:"interfaces,omitempty"`

	// Provides optional NFS rules
	// +kubebuilder:validation:Optional
	NfsRules []NfsRule `json:"rules,omitempty"`
}

type NfsRule struct {
	// Provides required NFS rule client match
	// +kubebuilder:validation:Required
	NfsRuleClient string `json:"client"`

	// Provides required NFS rule protocol
	// +kubebuilder:validation:Required
	NfsRuleProtocol string `json:"protocol"`

	// Provides required NFS rule read-write
	// +kubebuilder:validation:Optional
	NfsRuleRw string `json:"rw,omitempty"`

	// Provides required NFS rule read-only
	// +kubebuilder:validation:Optional
	NfsRuleRo string `json:"ro,omitempty"`

	// Provides required NFS rule superuser
	// +kubebuilder:validation:Optional
	NfsRuleSuperuser string `json:"superuser,omitempty"`

	// Provides required NFS rule anonyomous user
	// +kubebuilder:validation:Optional
	NfsRuleAnon string `json:"anon,omitempty"`
}
