package v1beta2

type NfsSubSpec struct {

	// Provides required NFS enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

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
	Lifs []LIF `json:"interfaces,omitempty"`

	// Provides optional NFS export definition
	// +kubebuilder:validation:Optional
	Export *NfsExport `json:"export,omitempty"`
}

type NfsExport struct {
	// Provides required NFS export name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Provides optional NFS rules
	// +kubebuilder:validation:Optional
	Rules []NfsRule `json:"rules,omitempty"`
}

type NfsRule struct {
	//Simplifed NFS Rule for now - but it should be like this
	/* 	// Provides required NFS rule client match
	   	// +kubebuilder:validation:Required
	   	Clients []NFSRuleClient `json:"clients"`

	   	// Provides required NFS rule protocol
	   	// +kubebuilder:validation:Required
	   	Protocols []NFSRuleVal `json:"protocols"`

	   	// Provides required NFS rule read-write
	   	// +kubebuilder:validation:Optional
	   	Rw []NFSRuleVal `json:"rw,omitempty"`

	   	// Provides required NFS rule read-only
	   	// +kubebuilder:validation:Optional
	   	Ro []NFSRuleVal `json:"ro,omitempty"`

	   	// Provides required NFS rule superuser
	   	// +kubebuilder:validation:Optional
	   	Superuser []NFSRuleVal `json:"superuser,omitempty"` */

	// Provides required NFS rule client match
	// +kubebuilder:validation:Required
	Clients string `json:"clients"`

	// Provides required NFS rule protocol
	// +kubebuilder:validation:Required
	Protocols string `json:"protocols"`

	// Provides required NFS rule read-write
	// +kubebuilder:validation:Optional
	Rw string `json:"rw,omitempty"`

	// Provides required NFS rule read-only
	// +kubebuilder:validation:Optional
	Ro string `json:"ro,omitempty"`

	// Provides required NFS rule superuser
	// +kubebuilder:validation:Optional
	Superuser string `json:"superuser,omitempty"`

	// Provides required NFS rule anonyomous user
	// +kubebuilder:validation:Optional
	Anon string `json:"anon,omitempty"`
}

// type NFSRuleVal struct {
// 	Value string
// }

// type NFSRuleClient struct {
// 	Match string `json:"match,omitempty"`
// }
