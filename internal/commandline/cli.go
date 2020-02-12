package commandline

import (
	"fmt"
	"io"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/ni/systemlink-cli/internal/model"
)

const profileFlag = "profile"
const verboseFlag = "verbose"
const apiKeyFlag = "api-key"
const usernameFlag = "username"
const passwordFlag = "password"
const urlFlag = "url"
const insecureFlag = "insecure"
const sshProxyFlag = "ssh-proxy"
const sshKeyFlag = "ssh-key"
const sshKnownHost = "ssh-known-host"

var globalFlags = []string{profileFlag, verboseFlag, apiKeyFlag, usernameFlag, passwordFlag, urlFlag, insecureFlag, sshProxyFlag, sshKeyFlag, sshKnownHost}

// CLI : The command line interface struct
type CLI struct {
	Parser    Parser
	Service   ServiceCaller
	Writer    io.Writer
	ErrWriter io.Writer
	Config    Config
}

func (c CLI) contains(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func (c CLI) isGlobalFlag(flag string) bool {
	return c.contains(flag, globalFlags)
}

func (c CLI) buildGlobalFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        apiKeyFlag,
			Usage:       "API KEY for accessing the NI services",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_API_KEY"},
			Hidden:      hidden,
		},
		&cli.StringFlag{
			Name:        usernameFlag,
			Usage:       "Username for basic auth (SystemLink Server)",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_USERNAME"},
			Hidden:      hidden,
		},
		&cli.StringFlag{
			Name:        passwordFlag,
			Usage:       "Password for basic auth (SystemLink Server)",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_PASSWORD"},
			Hidden:      hidden,
		},
		&cli.BoolFlag{
			Name:        verboseFlag,
			Usage:       "Provides debug output",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_VERBOSE"},
			Value:       false,
			Hidden:      hidden,
		},
		&cli.StringFlag{
			Name:        urlFlag,
			Usage:       "SystemLink server base URL",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_URL"},
			Hidden:      hidden,
		},
		&cli.StringFlag{
			Name:        profileFlag,
			Usage:       "Profile to load from configuration file",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_PROFILE"},
			Hidden:      hidden,
		},
		&cli.BoolFlag{
			Name:        insecureFlag,
			Usage:       "Ignore SSL certificate errors",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_INSECURE"},
			Value:       false,
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        sshProxyFlag,
			Usage:       "Use HTTP(S) over SSH",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_SSH_PROXY"},
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        sshKeyFlag,
			Usage:       "SSH private key used for authentication",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_SSH_KEY"},
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        sshKnownHost,
			Usage:       "Known host used for SSH host key auth",
			DefaultText: "using environment variable",
			EnvVars:     []string{"NI_SSH_KNOWN_HOST"},
			Hidden:      true,
		},
	}
}

func (c CLI) buildFlag(parameter model.Parameter) cli.Flag {
	return &cli.StringFlag{
		Name:  parameter.Name,
		Usage: parameter.Description,
	}
}

func uniqueFlags(flags []cli.Flag) []cli.Flag {
	keys := make(map[string]bool)
	list := []cli.Flag{}
	for _, flag := range flags {
		for _, name := range flag.Names() {
			if _, value := keys[name]; !value {
				keys[name] = true
				list = append(list, flag)
			}
		}
	}
	return list
}

func (c CLI) buildFlags(parameters []model.Parameter) []cli.Flag {
	flags := make([]cli.Flag, len(parameters))
	for i, p := range parameters {
		var flag = c.buildFlag(p)
		flags[i] = flag
	}
	return uniqueFlags(flags)
}

func (c CLI) validateRequiredFlags(context *cli.Context, parameters []model.Parameter) bool {
	var result = true
	for _, p := range parameters {
		if p.Required && !c.contains(p.Name, context.FlagNames()) {
			fmt.Fprintf(c.ErrWriter, "Missing argument: --%s\n", p.Name)
			result = false
		}
	}
	return result
}

func (c CLI) getFlagValues(context *cli.Context) map[string]string {
	var values = make(map[string]string)
	for _, p := range context.FlagNames() {
		if !c.isGlobalFlag(p) {
			values[p] = context.String(p)
		}
	}
	return values
}

func (c CLI) getSettings(context *cli.Context) model.Settings {
	profile := context.String(profileFlag)
	settings := c.Config.GetSettings(profile)

	if context.IsSet(urlFlag) {
		settings.URL = context.String(urlFlag)
	}
	if context.IsSet(apiKeyFlag) {
		settings.APIKey = context.String(apiKeyFlag)
	}
	if context.IsSet(usernameFlag) {
		settings.Username = context.String(usernameFlag)
	}
	if context.IsSet(passwordFlag) {
		settings.Password = context.String(passwordFlag)
	}
	if context.IsSet(verboseFlag) {
		settings.Verbose = context.Bool(verboseFlag)
	}
	if context.IsSet(insecureFlag) {
		settings.Insecure = context.Bool(insecureFlag)
	}
	if context.IsSet(sshProxyFlag) {
		settings.SSHProxy = context.String(sshProxyFlag)
	}
	if context.IsSet(sshKeyFlag) {
		settings.SSHKey = context.String(sshKeyFlag)
	}
	if context.IsSet(sshKnownHost) {
		settings.SSHKnownHost = context.String(sshKnownHost)
	}

	return settings
}

func (c CLI) buildSubCommand(definition model.Definition, operation model.Operation) *cli.Command {
	flags := c.buildFlags(operation.Parameters)

	return &cli.Command{
		Name:  operation.Name,
		Usage: operation.Description,
		Flags: append(flags, c.buildGlobalFlags(true)...),
		Action: func(context *cli.Context) error {
			if !c.validateRequiredFlags(context, operation.Parameters) {
				return nil
			}

			settings := c.getSettings(context)
			if settings.URL == "" {
				settings.URL = definition.URL
			}

			values := c.getFlagValues(context)
			parameterValues, err := ValueConverter{}.ConvertValues(values, operation.Parameters)
			if err != nil {
				fmt.Fprintln(c.ErrWriter, err)
				return nil
			}

			_, body, err := c.Service.Call(operation, parameterValues, settings)
			if err != nil {
				fmt.Fprintln(c.ErrWriter, err)
				return nil
			}

			fmt.Fprintln(c.Writer, body)
			return nil
		},
	}
}

func (c CLI) buildSubCommands(definition model.Definition, operations []model.Operation) []*cli.Command {
	commands := make([]*cli.Command, len(operations))
	for i, o := range operations {
		commands[i] = c.buildSubCommand(definition, o)
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	return commands
}

func (c CLI) buildCommand(definition model.Definition) *cli.Command {
	var subCommands = c.buildSubCommands(definition, definition.Operations)
	return &cli.Command{
		Name:        definition.Name,
		Subcommands: subCommands,
	}
}

func (c CLI) buildCommands(definitions []model.Definition) []*cli.Command {
	commands := make([]*cli.Command, len(definitions))
	for i, e := range definitions {
		commands[i] = c.buildCommand(e)
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	return commands
}

// Exec : Parses the given API models, validates and executes
// the given command line arguments
func (c CLI) Exec(args []string, models []model.Data) (*cli.App, int) {
	definitions, err := c.Parser.Parse(models)
	if err != nil {
		fmt.Fprintln(c.ErrWriter, err)
		return nil, 1
	}
	commands := c.buildCommands(definitions)

	app := &cli.App{
		Name:      "systemlink",
		Usage:     "Command-Line Interface for NI SystemLink Services",
		UsageText: "systemlink command [options]",
		Version:   "0.1.0",
		Commands:  commands,
		Flags:     c.buildGlobalFlags(false),
		Writer:    c.Writer,
		ErrWriter: c.ErrWriter,
	}

	app.Run(args)
	return app, 0
}
