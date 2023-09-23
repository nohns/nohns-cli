package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/nohns/nohns-cli/flag"
	"github.com/nohns/nohns-cli/subcmd"
	"github.com/urfave/cli/v2"
)

const (
	confDirName = ".nohns"
)

func main() {

	// Depedencies
	nvf := subcmd.NewNginxVhostFactory()

	app := cli.App{
		Name:                 "nohns",
		Usage:                "A simple CLI for automating repetivite tasks",
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			// Set home directory on context
			usr, err := user.Current()
			if err != nil {
				return fmt.Errorf("could not get current user: %s", err)
			}
			if err := c.Set(flag.NameHomeDir, usr.HomeDir); err != nil {
				return fmt.Errorf("could not set home directory on CLI context: %s", err)
			}

			// Make sure config directory exists
			if err := ensureConfDirIn(usr.HomeDir); err != nil {
				return err
			}

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flag.NameHomeDir,
				Usage: "The home directory of the current user",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "cp-pubkey",
				Aliases: []string{"pk"},
				Usage:   "Copies public key to clipboard",
				Action:  subcmd.ActionCopyPubSSHKey,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  flag.NameSSHKey,
						Usage: "The name of the public key file",
						Value: flag.DefaultSSHKey,
					},
				},
			},
			{
				Name:        "nginx",
				Aliases:     []string{"i"},
				Usage:       "Run nginx related tasks",
				Subcommands: subcmd.NginxSubcommands(nvf),
			},
		},
	}

	// Run cli
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\nerror: %s\n", err)
		os.Exit(1)
	}
}

// Ensures that a config directory exists
func ensureConfDirIn(basepath string) error {
	p := filepath.Join(basepath, confDirName)
	fi, err := os.Stat(p)
	if os.IsNotExist(err) {
		if err := os.Mkdir(p, 0700); err != nil {
			return fmt.Errorf("could not create %s: %s", p, err)
		}

		return nil
	}
	if err != nil {
		return fmt.Errorf("could not create %s: %s", p, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory. please remove it and try running the cli again.", p)
	}

	return nil
}
