package main

import (
	"fmt"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go/pkg/args"
)

var envCmd = &cmd.Command{
	Name: "env",
	Help: "manage environments, credentials, and base URLs",
	Subcommands: []*cmd.Command{
		envListCmd,
		envCreateCmd,
		envSetCmd,
		envDeleteCmd,
		envKeyCmd,
		envURLCmd,
	},
}

var envListCmd = &cmd.Command{
	Name: "list",
	Help: "list environments",
	Handler: func(i *cmd.Input) error {

		cfg, err := loadConfig(i)
		if err != nil {
			return err
		}

		active := cfg.ActiveEnv
		for name := range cfg.Envs {
			marker := " "
			if name == active {
				marker = "*"
			}
			fmt.Printf("%s %s\n", marker, name)
		}

		return nil
	},
}

var envCreateCmd = &cmd.Command{
	Name: "create",
	Help: "create environment",
	Operands: []cmd.Operand{
		{
			Name: "name",
			Help: "environment name",
		},
	},
	Options: []cmd.Option{
		{
			Long: "base-url",
			Type: cmd.OptionTypeParameter,
			Help: "set base url",
		},
		{
			Long: "api-key",
			Type: cmd.OptionTypeParameter,
			Help: "set api key",
		},
		{
			Long: "bootstrap",
			Type: cmd.OptionTypeFlag,
			Help: "generate new api key",
		},
	},
	Handler: func(i *cmd.Input) error {

		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("<name> is empty")
		}

		cfg, err := loadConfig(i)
		if err != nil {
			return err
		}

		envCfg := cfg.Envs[name]
		if url := i.GetParameter("base-url"); url != nil && *url != "" {
			envCfg.BaseURL = strings.TrimRight(*url, "/")
		}

		if i.GetFlag("bootstrap") {

			// generate a brand new API key for this env
			key, err := generateAPIKey()
			if err != nil {
				return err
			}

			// write key to stdout, so shell can do something with it
			fmt.Print(key)

			// set same API key to the newly created environment
			envCfg.APIKey = key

		} else {
			// if not bootstrapping, look for passed API key
			key := i.GetParameter("api-key")
			if key != nil && *key != "" {
				envCfg.APIKey = *key
			}
		}

		if cfg.Envs == nil {
			cfg.Envs = map[string]EnvConfig{}
		}
		cfg.Envs[name] = envCfg

		return saveConfig(i, cfg)
	},
}

var envSetCmd = &cmd.Command{
	Name: "set",
	Help: "set active environment",
	Operands: []cmd.Operand{
		{
			Name: "name",
			Help: "environment name",
		},
	},
	Handler: func(i *cmd.Input) error {

		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("<name> is empty")
		}

		return saveActiveEnv(i, name)
	},
}

var envDeleteCmd = &cmd.Command{
	Name: "delete",
	Help: "delete environment",
	Operands: []cmd.Operand{
		{
			Name: "name",
			Help: "environment name",
		},
	},
	Handler: func(i *cmd.Input) error {

		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("<name> is empty")
		}

		cfg, err := loadConfig(i)
		if err != nil {
			return err
		}

		delete(cfg.Envs, name)
		if cfg.ActiveEnv == name {
			cfg.ActiveEnv = "default"
		}

		return saveConfig(i, cfg)
	},
}

var envKeyCmd = &cmd.Command{
	Name: "key",
	Help: "manage stored api key for active environment",
	Subcommands: []*cmd.Command{
		envKeySetCmd,
		envKeyClearCmd,
	},
}

var envKeySetCmd = &cmd.Command{
	Name: "set",
	Help: "store provided api key",
	Operands: []cmd.Operand{
		{
			Name: "key",
			Help: "api key token",
		},
	},
	Handler: func(i *cmd.Input) error {

		key := i.GetOperand("key")
		if key == "" {
			return fmt.Errorf("<key> is empty")
		}

		return saveAPIKey(i, key)
	},
}

var envKeyClearCmd = &cmd.Command{
	Name: "clear",
	Help: "remove saved api key",
	Handler: func(i *cmd.Input) error {
		return deleteAPIKey(i)
	},
}

var envURLCmd = &cmd.Command{
	Name: "url",
	Help: "manage api base url for active environment",
	Subcommands: []*cmd.Command{
		envURLGetCmd,
		envURLSetCmd,
		envURLClearCmd,
	},
}

var envURLGetCmd = &cmd.Command{
	Name: "get",
	Help: "print base url",
	Handler: func(i *cmd.Input) error {

		url, err := loadBaseURL(i)
		if err != nil {
			return err
		}

		if url == "" {
			fmt.Println("none set")
		} else {
			fmt.Println(url)
		}

		return nil
	},
}

var envURLSetCmd = &cmd.Command{
	Name: "set",
	Help: "set base url",
	Operands: []cmd.Operand{
		{
			Name: "url",
			Help: "base url",
		},
	},
	Handler: func(i *cmd.Input) error {

		u := i.GetOperand("url")
		if u == "" {
			return fmt.Errorf("<url> is empty")
		}

		return saveBaseURL(i, strings.TrimRight(u, "/"))
	},
}

var envURLClearCmd = &cmd.Command{
	Name: "clear",
	Help: "clear base url",
	Handler: func(i *cmd.Input) error {
		return deleteBaseURL(i)
	},
}
