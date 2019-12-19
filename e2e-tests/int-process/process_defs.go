package intprocess

import (
	"os/exec"

	"github.com/mobiledgex/edge-cloud/integration/process"
)

type MC struct {
	process.Common `yaml:",inline"`
	Addr           string
	SqlAddr        string
	VaultAddr      string
	RolesFile      string
	TLS            process.TLSCerts
	cmd            *exec.Cmd
}
type Sql struct {
	process.Common `yaml:",inline"`
	DataDir        string
	HttpAddr       string
	Username       string
	Dbname         string
	TLS            process.TLSCerts
	cmd            *exec.Cmd
}
type Shepherd struct {
	process.Common `yaml:",inline"`
	NotifyAddrs    string
	Interval       string
	Platform       string
	VaultAddr      string
	PhysicalName   string
	CloudletKey    string
	TLS            process.TLSCerts
	cmd            *exec.Cmd
	Span           string
	Region         string
}
type AutoProv struct {
	process.Common `yaml:",inline"`
	NotifyAddrs    string
	CtrlAddrs      string
	VaultAddr      string
	InfluxAddr     string
	TLS            process.TLSCerts
	ShortTimeouts  bool
	cmd            *exec.Cmd
}

type PromE2e struct {
	process.Common `yaml:",inline"`
	Port           int
	cmd            *exec.Cmd
}

type Exporter struct {
	process.Common `yaml:",inline"`
	DataFile       string
	Port           int
	cmd            *exec.Cmd
}
