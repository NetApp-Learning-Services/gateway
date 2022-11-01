package ontap

type NFSCreation struct {
	Enabled  bool        `json:"enabled,omitempty"`
	Protocol NFSProtocol `json:"protocol,omitempty"`
	Svm      NfsSvm      `json:"svm,omitempty"`
}

type NFSProtocol struct {
	V3Enable  bool `json:"v3_enabled,omitempty"`
	V4Enable  bool `json:"v40_enabled,omitempty"`
	V41Enable bool `json:"v41_enabled,omitempty"`
}

type NfsSvm struct {
	Name string `json:"name,omitempty"`
	Uuid string `json:"uuid,omitempty"`
}

type ExportPolicyCreation struct {
	Name  string       `json:"name,omitempty"`
	Svm   NfsSvm       `json:"svm,omitempty"`
	Rules []ExportRule `json:"rules,omitempty"`
}

// Requires ONTAP 9.10?
type ExportRule struct {
	Protocols string `json:"protocols,omitempty"`
	RwRule    string `json:"rw_rule,omitempty"`
	RoRule    string `json:"ro_rule,omitempty"`
	Superuser string `json:"superuser,omitempty"`
	Anonuser  string `json:"anonymous_user,omitempty"`
}
