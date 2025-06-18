package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var confPath string

var ProxyCommand = &cli.Command{
	Name:  "proxy",
	Usage: "Manage nginx proxy configurations",
	Subcommands: []*cli.Command{
		{
			Name:  "add",
			Usage: "Add new nginx proxy configuration",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "conf",
					Usage:       "Path to the nginx configuration file",
					Value:       "nginx.conf",
					Destination: &confPath,
				},
			},
			Action: func(c *cli.Context) error {
				appName := c.Args().First()
				if appName == "" {
					return fmt.Errorf("❌ You must provide app name, e.g. `zeroops proxy add my-app --conf nginx.conf`")
				}

				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}

				fmt.Printf("Deploying nginx configuration for %s...\n", appName)

				if _, err := os.Stat(confPath); os.IsNotExist(err) {
					return fmt.Errorf("❌ File %s does not exist", confPath)
				}

				if err := deployNginxConf(ctx, appName, confPath); err != nil {
					return fmt.Errorf("❌ Failed to deploy nginx configuration: %w", err)
				}

				fmt.Printf("✅ Nginx configuration for %s deployed and reloaded successfully\n", appName)
				return nil
			},
		},
		{
			Name:  "status",
			Usage: "Show nginx config for an app on the VPS",
			Action: func(c *cli.Context) error {
				appName := c.Args().First()
				if appName == "" {
					return fmt.Errorf("❌ You must provide app name, e.g. `zeroops proxy status my-app`")
				}
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}
				cmd := fmt.Sprintf("cat /etc/nginx/sites-enabled/%s", appName)
				return runSSH(ctx, cmd)
			},
		},
		{
			Name:  "list",
			Usage: "List all nginx configs on the VPS",
			Action: func(c *cli.Context) error {
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}
				cmd := "ls -1 /etc/nginx/sites-enabled"
				return runSSH(ctx, cmd)
			},
		},
		{
			Name:  "rm",
			Usage: "Remove nginx config for an app and reload nginx",
			Action: func(c *cli.Context) error {
				appName := c.Args().First()
				if appName == "" {
					return fmt.Errorf("❌ You must provide app name, e.g. `zeroops proxy rm my-app`")
				}
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}
				cmd := fmt.Sprintf("sudo rm -f /etc/nginx/sites-enabled/%s && sudo nginx -s reload", appName)
				return runSSH(ctx, cmd)
			},
		},
	},
}

// deployNginxConf copies a local nginx config file to /tmp, then moves it to /etc/nginx/sites-enabled/appName.conf and reloads nginx.
func deployNginxConf(ctx *Context, appName, localConfPath string) error {
	remoteTmpPath := "/tmp/" + appName
	targetPath := "/etc/nginx/sites-enabled/" + appName

	fmt.Printf("📁 Copying configuration to VPS...\n")
	// Copy to VPS /tmp
	scpCmd := exec.Command("scp", localConfPath, fmt.Sprintf("%s@%s:%s", ctx.User, ctx.Host, remoteTmpPath))
	scpCmd.Stdout = nil
	scpCmd.Stderr = nil
	if err := scpCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy config to VPS: %w", err)
	}

	fmt.Printf("🚀 Moving configuration to nginx sites-enabled...\n")
	// Move to /etc/nginx/sites-enabled using SSH
	mvCmd := fmt.Sprintf("sudo mv %s %s", remoteTmpPath, targetPath)
	if err := runSSH(ctx, mvCmd); err != nil {
		return fmt.Errorf("failed to move config to sites-enabled: %w", err)
	}

	fmt.Printf("🔄 Reloading nginx configuration...\n")
	// Reload nginx using SSH
	reloadCmd := "sudo nginx -s reload"
	if err := runSSH(ctx, reloadCmd); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}
