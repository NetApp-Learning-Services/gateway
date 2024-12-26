package v1beta2

type S3SubSpec struct {

	// Provides required S3 enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

	// Provides required S3 server name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Provides optional S3 LIFs
	// +kubebuilder:validation:Optional
	Lifs []LIF `json:"interfaces,omitempty"`

	// Provides optional user definition
	// +kubebuilder:validation:Optional
	Users []S3User `json:"users,omitempty"`

	// Provides optional Http definition
	// +kubebuilder:validation:Optional
	Http *S3Http `json:"http,omitempty"`

	// Provides optional Https definition
	// +kubebuilder:validation:Optional
	Https *S3Https `json:"https,omitempty"`

	// Provides optional buckets definition
	// +kubebuilder:validation:Optional
	Buckets []S3Bucket `json:"buckets,omitempty"`
}

type S3User struct {
	// Provides required user name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Provides optional namespace
	//+kubebuilder:validation:Optional
	Namespace *string `json:"namespace,omitempty"`
}

type S3Http struct {
	// Provides required S3 http enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

	// Provides required S3 http enablement
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=80
	Port int `json:"port"`
}

type S3Https struct {
	// Provides required S3 https enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

	// Provides required S3 https enablement
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=443
	Port int `json:"port"`
}

type S3Tls struct {
	// Provides required S3 tls enablement
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`
}

type S3Bucket struct {
	// Provides required S3 bucket name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Provides required S3 bucket vsadmin managementability
	// +kubebuilder:validation:Optional
	Allowed bool `json:"allowed,omitempty"`

	// Provides required S3 bucket size
	// +kubebuilder:validation:Optional
	Size int `json:"size,omitempty"`

	// Provides required S3 bucket type
	// +kubebuilder:validation:Optional
	Type string `json:"type,omitempty"`
}
