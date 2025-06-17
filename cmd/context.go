package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

var ContextCommand = &cli.Command{
	Name:  "context",
	Usage: "Manage VPS contexts",
	Subcommands: []*cli.Command{
		{
			Name:      "add",
			Usage:     "Add new VPS context like Docker",
			ArgsUsage: "[name]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "docker",
					Usage:    "SSH URI like: host=ssh://user@host",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				name := c.Args().First()
				if name == "" {
					return fmt.Errorf("❌ context name is required, e.g. zeroops context add my-vps --docker \"host=ssh://user@ip\"")
				}

				raw := c.String("docker")
				if !strings.HasPrefix(raw, "host=ssh://") {
					return fmt.Errorf("❌ invalid format: must be like host=ssh://user@ip")
				}

				parts := strings.SplitN(strings.TrimPrefix(raw, "host=ssh://"), "@", 2)
				if len(parts) != 2 {
					return fmt.Errorf("❌ invalid ssh URI, expected user@host")
				}
				user := parts[0]
				host := parts[1]

				home, _ := os.UserHomeDir()
				ctxPath := filepath.Join(home, ".zeroops", "contexts", name+".yaml")
				os.MkdirAll(filepath.Dir(ctxPath), 0700)

				content := fmt.Sprintf("host: %s\nuser: %s\n", host, user)
				if err := os.WriteFile(ctxPath, []byte(content), 0600); err != nil {
					return fmt.Errorf("❌ failed to save context: %w", err)
				}

				fmt.Printf("✅ Context '%s' saved: %s@%s\n", name, user, host)
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "List available contexts",
			Action: func(c *cli.Context) error {
				home, _ := os.UserHomeDir()
				ctxDir := filepath.Join(home, ".zeroops", "contexts")
				entries, _ := os.ReadDir(ctxDir)
				if len(entries) == 0 {
					fmt.Println("No contexts found")
					return nil
				}
				fmt.Println("Available contexts:")
				for _, entry := range entries {
					fmt.Println("- " + strings.TrimSuffix(entry.Name(), ".yaml"))
				}
				return nil
			},
		},
		{
			Name:  "use",
			Usage: "Set current context",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Required: true},
			},
			Action: func(c *cli.Context) error {
				home, _ := os.UserHomeDir()
				path := filepath.Join(home, ".zeroops", "current-context")
				return os.WriteFile(path, []byte(c.String("name")), 0600)
			},
		},
		{
			Name:  "current",
			Usage: "Show current context",
			Action: func(c *cli.Context) error {
				home, _ := os.UserHomeDir()
				path := filepath.Join(home, ".zeroops", "current-context")
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Println("No current context set")
					return nil
				}
				fmt.Println("Current context: " + string(content))
				return nil
			},
		},
		{
			Name:  "rm",
			Usage: "Remove context",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Required: true},
			},
			Action: func(c *cli.Context) error {
				home, _ := os.UserHomeDir()
				ctxDir := filepath.Join(home, ".zeroops", "contexts")
				ctxName := c.String("name")
				ctxPath := filepath.Join(ctxDir, ctxName+".yaml")
				os.Remove(ctxPath)

				currentPath := filepath.Join(home, ".zeroops", "current-context")
				content, err := os.ReadFile(currentPath)
				if err == nil && strings.TrimSpace(string(content)) == ctxName {
					os.Remove(currentPath)
				}
				fmt.Printf("✅ Context '%s' removed\n", ctxName)
				return nil
			},
		},
	},
}
