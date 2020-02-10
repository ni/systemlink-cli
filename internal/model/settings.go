package model

// Settings are all the global CLI input parameters which
// can be provided through global CLI switches or the
// systemlink.yaml configuration file
type Settings struct {
	APIKey       string
	Username     string
	Password     string
	Verbose      bool
	URL          string
	Insecure     bool
	SSHProxy     string
	SSHKey       string
	SSHKnownHost string
}
