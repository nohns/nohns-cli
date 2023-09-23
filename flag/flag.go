package flag

import (
	"github.com/urfave/cli/v2"
)

const (
	NameHomeDir = "home_dir"
)

func HomeDir(c *cli.Context) string {
	return c.String(NameHomeDir)
}
