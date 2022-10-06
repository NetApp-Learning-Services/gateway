package v1alpha1

// NamespacedName contains the name of a object and its namespace
type NamespacedName struct {

	// Provides optional namespace
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// Provides credentials name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	Name string `json:"name"`
}

/// IPFormat Regex to support both IPV4 and IPV6 format
/// +kubebuilder:validation:Pattern="((^((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))$)|(^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$))"
///type IPFormat string

// ManagementLIF contains parameters regarding the SVM's management LIF
type ManagementLIF struct {

	// Provides Management LIF name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	Name string `json:"name"`

	// Provides Management LIF IP address
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern:=ip
	IPAddress string `json:"ip"`

	// Provides Management LIF netmask
	// +kubebuilder:validation:Required
	Netmask string `json:"netmask"`

	// Provides Management LIF broadcast domain
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	BroacastDomain string `json:"broacastDomain"`

	// Provides Management LIF home node
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	HomeNode string `json:"homeNode"`
}

// OperationState defines the potential states
type OperationState string

// SvmOperationStates defined here
const (
	SvmOperationStateProvisioning OperationState = "Provisioning"
	SvmOperationStateProvisioned  OperationState = "Provisioned"
	SvmOperationStateFailed       OperationState = "Failed"
	SvmOperationStateUnknown      OperationState = "Unknown"
	SvmOperationStateDeleting     OperationState = "Deleting"
)
