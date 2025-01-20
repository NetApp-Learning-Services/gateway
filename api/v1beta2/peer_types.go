package v1beta2

type PeerSubSpec struct {

	// Provides required peering relationship name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	Name string `json:"name"`

	// Provides required peering relationship passphrase
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	Passphrase string `json:"passphrase"`

	// Provides required peering encryption
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum="none";"tls-psk"
	// +kubebuilder:default:=tls-psk
	Encryption string `json:"encryption"`

	// Provides optional peer applications
	// +kubebuilder:validation:Optional
	Applications []PeerApplication `json:"applications,omitempty"`

	// Provides optional remote peer cluster
	// +kubebuilder:validation:Optional
	Remote PeerRemote `json:"remote,omitempty"`

	// Provides optional intercluster LIFs
	// +kubebuilder:validation:Optional
	Lifs []LIF `json:"interfaces,omitempty"`
}

type PeerApplication struct {
	// Provides required peering application
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum="snapmirror";"flexcache"
	App string `json:"app"`
}

type PeerRemote struct {
	// Provides optional remote cluster name
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Format:=string
	Clustername string `json:"clusterName,omitempty"`

	// Provides required remote cluster intercluster LIF
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`((^\s*((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))\s*$)|(^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$))`
	Ipaddress string `json:"ipAddress,omitempty"`

	// Provides required remote svm name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	Svmname string `json:"svmName,omitempty"`
}
