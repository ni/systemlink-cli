package commandline

import (
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/ni/systemlink-cli/internal/model"
)

// Config contains the parsed YAML data from the systemlink.yaml
// configuration file
type Config struct {
	Profiles []profile `yaml:"profiles"`
}

type profile struct {
	Name         string `yaml:"name"`
	APIKey       string `yaml:"api-key"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Verbose      bool   `yaml:"verbose"`
	URL          string `yaml:"url"`
	Insecure     bool   `yaml:"insecure"`
	SSHProxy     string `yaml:"ssh-proxy"`
	SSHKey       string `yaml:"ssh-key"`
	SSHKnownHost string `yaml:"ssh-known-host"`
}

func (c *Config) resolveRelativePath(path string, baseDir string) string {
	if path == "" || filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
		return filepath.FromSlash(path)
	}
	return filepath.Join(baseDir, path)
}

func (c *Config) resolvePaths(baseDir string) {
	var result []profile
	for _, p := range c.Profiles {
		p.SSHKey = c.resolveRelativePath(p.SSHKey, baseDir)
		result = append(result, p)
	}
	c.Profiles = result
}

// NewConfig initializes a new Config structure based on the yaml data
// in the given byte stream
func NewConfig(input []byte, baseDir string) Config {
	c := Config{}
	err := yaml.Unmarshal(input, &c)
	c.resolvePaths(baseDir)
	if err != nil {
		panic("Error loading configuration file: " + err.Error())
	}
	return c
}

func (c *Config) findProfile(profileName string) profile {
	if profileName == "" {
		profileName = "default"
	}

	for _, p := range c.Profiles {
		if p.Name == profileName {
			return p
		}
	}
	return profile{}
}

// GetSettings returns the Settings structure with the data from the
// selected profile of the yaml configuration file
func (c *Config) GetSettings(profileName string) model.Settings {
	profile := c.findProfile(profileName)

	return model.Settings{
		APIKey:       profile.APIKey,
		Username:     profile.Username,
		Password:     profile.Password,
		Verbose:      profile.Verbose,
		URL:          profile.URL,
		Insecure:     profile.Insecure,
		SSHProxy:     profile.SSHProxy,
		SSHKey:       profile.SSHKey,
		SSHKnownHost: profile.SSHKnownHost,
	}
}
