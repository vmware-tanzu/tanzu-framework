package template

type VCenterClusterVar struct {
	DataCenter    string `json:"datacenter"`
	Server        string `json:"server"`
	Template      string `json:"template"`
	TLSThumbprint string `json:"tlsThumbprint"`
}
