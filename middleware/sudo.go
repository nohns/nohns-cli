package middleware

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func RequireSudo(f cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		if os.Geteuid() != 0 {
			return fmt.Errorf("this command requires sudo.")
		}

		if f != nil {
			return f(c)
		}

		return nil
	}
}
