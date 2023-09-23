package subcmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/nohns/nohns-cli/flag"
	"github.com/urfave/cli/v2"
)

// Copies public SSH key to clipboard
func ActionCopyPubSSHKey(c *cli.Context) error {
	keyname := flag.SSHKey(c)

	p := filepath.Join(flag.HomeDir(c), ".ssh", keyname)
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("could not open public SSH key (%s) on disk: %s", p, err)
	}

	pubkey, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("could not read public SSH (%s) from disk: %s", p, err)
	}
	clipboard.WriteAll(string(pubkey))
	fmt.Printf("Copied public SSH key to clipboard!\n")

	return nil
}
