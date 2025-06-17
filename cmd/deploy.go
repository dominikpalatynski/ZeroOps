package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
)

type Context struct {
	Host string
	User string
	Key  string
}

var DeployCommand = &cli.Command{
	Name:  "deploy",
	Usage: "Manage deployments",
	Subcommands: []*cli.Command{
		{
			Name:  "add",
			Usage: "Render docker-compose, send to VPS, and deploy",
			Action: func(c *cli.Context) error {
				appName := c.Args().First()
				if appName == "" {
					return fmt.Errorf("❌ You must provide app name, e.g. `zeroops deploy add my-app`")
				}
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}

				remotePath := fmt.Sprintf("/apps/%s", appName)

				env, err := loadEnv(".env")
				if err != nil {
					return fmt.Errorf("❌ Failed to read .env: %w", err)
				}

				rendered, err := renderTemplate("docker-compose.tpl.yml", env)
				if err != nil {
					return fmt.Errorf("❌ Failed to render template: %w", err)
				}

				if err := os.WriteFile("docker-compose.yml", []byte(rendered), 0644); err != nil {
					return fmt.Errorf("❌ Failed to write docker-compose.yml: %w", err)
				}

				fmt.Println("📁 Ensuring remote directory exists...")
				fmt.Printf("ssh %s@%s sudo mkdir -p %s\n", ctx.User, ctx.Host, remotePath)
				mkdirCmd := fmt.Sprintf("sudo mkdir -p %s", remotePath)
				sshMkdir := exec.Command("ssh", fmt.Sprintf("%s@%s", ctx.User, ctx.Host), mkdirCmd)
				sshMkdir.Stdout = os.Stdout
				sshMkdir.Stderr = os.Stderr
				if err := sshMkdir.Run(); err != nil {
					return fmt.Errorf("❌ Failed to create remote directory: %w", err)
				}

				fmt.Println("🛁 Copying files to VPS...")
				if err := scpWithSudo(ctx, "docker-compose.yml", remotePath+"/docker-compose.yml"); err != nil {
					return err
				}
				if err := scpWithSudo(ctx, ".env", remotePath+"/.env"); err != nil {
					return err
				}

				fmt.Println("🚀 Running docker-compose on VPS...")
				sshCmd := fmt.Sprintf("cd %s && sudo docker compose up --build -d", remotePath)
				cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", ctx.User, ctx.Host), sshCmd)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					// Cleanup if docker build fails
					fmt.Println("❌ Docker build failed, cleaning up remote app directory...")
					cleanupCmd := fmt.Sprintf("sudo rm -rf %s", remotePath)
					cleanup := exec.Command("ssh", fmt.Sprintf("%s@%s", ctx.User, ctx.Host), cleanupCmd)
					cleanup.Stdout = os.Stdout
					cleanup.Stderr = os.Stderr
					cleanup.Run()
					return fmt.Errorf("❌ Docker build failed: %w", err)
				}

				fmt.Println("✅ Deployment complete!")
				return nil
			},
		},
		{
			Name:  "status",
			Usage: "Check docker-compose status for an app",
			Action: func(c *cli.Context) error {
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}
				appName := c.Args().First()
				if appName == "" {
					return fmt.Errorf("❌ You must provide app name")
				}
				cmd := fmt.Sprintf("cd /apps/%s && sudo docker compose ps", appName)
				return runSSH(ctx, cmd)
			},
		},
		{
			Name:  "rm",
			Usage: "Remove app (containers and directory)",
			Action: func(c *cli.Context) error {
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}
				appName := c.Args().First()
				if appName == "" {
					return fmt.Errorf("❌ You must provide app name")
				}
				cmd := fmt.Sprintf("sudo docker compose -f /apps/%s/docker-compose.yml down && sudo rm -rf /apps/%s", appName, appName)
				return runSSH(ctx, cmd)
			},
		},
		{
			Name:  "list",
			Usage: "List all deployed apps",
			Action: func(c *cli.Context) error {
				ctx, err := loadCurrentContext()
				if err != nil {
					return err
				}
				cmd := "ls -1 /apps"
				return runSSH(ctx, cmd)
			},
		},
	},
}

func loadCurrentContext() (*Context, error) {
	usr, _ := user.Current()
	base := filepath.Join(usr.HomeDir, ".zeroops")

	data, err := os.ReadFile(filepath.Join(base, "current-context"))
	if err != nil {
		return nil, fmt.Errorf("❌ No current context selected")
	}

	name := strings.TrimSpace(string(data))
	content, err := os.ReadFile(filepath.Join(base, "contexts", name+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to read context: %w", err)
	}

	ctx := &Context{}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "host:") {
			ctx.Host = strings.TrimSpace(strings.TrimPrefix(line, "host:"))
		} else if strings.HasPrefix(line, "user:") {
			ctx.User = strings.TrimSpace(strings.TrimPrefix(line, "user:"))
		} else if strings.HasPrefix(line, "key:") {
			ctx.Key = strings.TrimSpace(strings.TrimPrefix(line, "key:"))
		}
	}

	return ctx, nil
}

func loadEnv(path string) (map[string]string, error) {
	env := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		env[key] = val
	}

	return env, scanner.Err()
}

func renderTemplate(templatePath string, data map[string]string) (string, error) {
	tplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("compose").Delims("{{", "}}").Parse(string(tplBytes))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func scpWithSudo(ctx *Context, localPath, remotePath string) error {
	tmpPath := "/tmp/" + filepath.Base(localPath)

	cmd := exec.Command("scp", localPath, fmt.Sprintf("%s@%s:%s", ctx.User, ctx.Host, tmpPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	moveCmd := exec.Command("ssh", fmt.Sprintf("%s@%s", ctx.User, ctx.Host),
		fmt.Sprintf("sudo mv %s %s", tmpPath, remotePath),
	)
	moveCmd.Stdout = os.Stdout
	moveCmd.Stderr = os.Stderr
	return moveCmd.Run()
}

func runSSH(ctx *Context, cmd string) error {
	ssh := exec.Command("ssh", fmt.Sprintf("%s@%s", ctx.User, ctx.Host), cmd)
	ssh.Stdout = os.Stdout
	ssh.Stderr = os.Stderr
	return ssh.Run()
}
