package subcmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/nohns/nohns-cli/middleware"
	"github.com/urfave/cli/v2"
)

const (
	nginxVhostAvailableDir = "/etc/nginx/sites-available"
	nginxVhostEnabledDir   = "/etc/nginx/sites-enabled"
)

const (
	nginxPhpFlagAlias     = "alias"
	nginxPhpFlagHandle404 = "handle-404"
	nginxPhpFlagDir       = "dir"
	nginxWWWDir           = "/var/www"
)

func Nginx(nvf *nginxVhostFactory) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "vhost",
			Before: cli.BeforeFunc(middleware.RequireSudo(nil)),
			Subcommands: []*cli.Command{
				{
					Name: "php",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  nginxPhpFlagHandle404,
							Value: true,
						},
						&cli.StringSliceFlag{
							Name:    nginxPhpFlagAlias,
							Aliases: []string{"a"},
							Usage:   "An alias for the hostname",
							Value:   cli.NewStringSlice(),
						},
						&cli.StringFlag{
							Name:  nginxPhpFlagDir,
							Usage: "The directory to serve",
						},
					},
					Action: func(c *cli.Context) error {
						if c.Args().Len() < 1 {
							return fmt.Errorf("expected 1 arguments: [hostname]")
						}
						hostname := c.Args().Get(0)

						// If no directory is specified, use the hostname as the directory
						rootDir := filepath.Join(nginxWWWDir, hostname)
						if c.IsSet(c.String(nginxPhpFlagDir)) {
							rootDir = c.String(nginxPhpFlagDir)
						}

						// Create nginx PHP vhost on disk
						err := nvf.createPHPHost(&nginxVhostConf{
							Hostname:    hostname,
							Handle404:   c.Bool(nginxPhpFlagHandle404),
							HostAliases: strings.Join(c.StringSlice(nginxPhpFlagAlias), " "),
							RootDir:     rootDir,
						})
						if err != nil {
							return fmt.Errorf("could not create PHP nginx vhost: %v", err)
						}

						// Create directory if it doesn't exist
						if err := os.Mkdir(rootDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
							return fmt.Errorf("could not create www root directory %s: %s", rootDir, err)
						}
						// Set www directory permissions
						if err := os.Chmod(rootDir, 0755); err != nil {
							return fmt.Errorf("could not set permissions on www root directory %s: %s", rootDir, err)
						}
						// Set www directory owner
						uid, err := strconv.Atoi(os.Getenv("SUDO_UID"))
						if err != nil {
							return fmt.Errorf("could not get SUDO_UID: %s", err)
						}
						gid, err := strconv.Atoi(os.Getenv("SUDO_GID"))
						if err != nil {
							return fmt.Errorf("could not get SUDO_GID: %s", err)
						}
						if err := os.Chown(rootDir, uid, gid); err != nil {
							return fmt.Errorf("could not set owner '%s' on www root directory %s: %s", os.Getenv("SUDO_USER"), rootDir, err)
						}

						// Ask user if they want to enable the vhost
						prompt := promptui.Prompt{
							Label:     "Do you want to enable this vhost?",
							IsConfirm: true,
						}
						if _, err = prompt.Run(); errors.Is(err, promptui.ErrAbort) {
							return nil
						}

						// Make available by symlinking to sites-enabled
						availablefp := filepath.Join(nginxVhostAvailableDir, hostname)
						enablefp := filepath.Join(nginxVhostEnabledDir, hostname)
						if err := os.Symlink(availablefp, enablefp); err != nil && !errors.Is(err, os.ErrExist) {
							return fmt.Errorf("could not enable vhost: %s", err)
						}
						fmt.Printf("added symbolic link for vhost to sites-enabled.\n")

						// Reload nginx if on linux
						if runtime.GOOS != "linux" {
							return fmt.Errorf("nginx reload only supported on linux (with systemctl)")
						}
						if err := exec.Command("systemctl", "nginx", "reload").Run(); err != nil {
							return fmt.Errorf("could not reload systemd nginx: %s", err)
						}

						return nil
					},
				},
			},
		},
	}
}

type nginxVhostFactory struct {
}

func NewNginxVhostFactory() *nginxVhostFactory {
	return &nginxVhostFactory{}
}

type nginxRevProxyConf struct {
	hostname string
	target   string
}

// Creates a new reverse proxy nginx vhost in the nginx sites-available directory
func (nvf *nginxVhostFactory) createRevProxy(conf *nginxRevProxyConf) error {
	return nil
}

// Creates a new PHP nginx vhost in the nginx sites-available directory
func (nvf *nginxVhostFactory) createPHPHost(conf *nginxVhostConf) error {
	conf.Body = `
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
    }

    location ~ /\.ht {
        deny all;
    }
	`

	if err := nvf.createHost(conf); err != nil {
		return err
	}

	return nil
}

type nginxVhostConf struct {
	Hostname    string
	HostAliases string
	RootDir     string
	Handle404   bool
	Body        string
}

func (nvf *nginxVhostFactory) createHost(conf *nginxVhostConf) error {
	// Open vhost file for writing template
	/*p := filepath.Join(nginxVhostAvailableDir, conf.hostname)
	_, err := os.OpenFile(p, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open %s for writing: %s", p, err)
	}*/

	// Parse vhost template
	t, err := template.New("nginxVhost").Parse(nginxVhostTemplate)
	if err != nil {
		return fmt.Errorf("could not parse nginx vhost template: %s", err)
	}

	// Render template with config and write to file
	if err := t.Execute(os.Stdout, conf); err != nil {
		return fmt.Errorf("could not write nginx vhost template to file: %s", err)
	}

	return nil
}

const nginxVhostTemplate = `# Auto-generated by nohns CLI (https://github.com/nohns/nohns-cli) - use 'nohns nginx vhost' to generate similar files
server {
    server_name {{ .Hostname }} {{ .HostAliases }};
    root {{ .RootDir }};
	{{- if .Handle404 }}

    location / {
        try_files $uri $uri/ =404;
    }
	{{ end -}}

	{{- .Body }}
    listen 80;
}

`
