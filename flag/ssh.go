package flag

import "github.com/urfave/cli/v2"

const (
	NameSSHKey    = "key"
	DefaultSSHKey = "id_rsa.pub"
)

func SSHKey(c *cli.Context) string {
	return c.String(NameSSHKey)
}
