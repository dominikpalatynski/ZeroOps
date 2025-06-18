package main

import (
	"log"
	"os"

	"github.com/yourname/zeroops/cmd"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "zeroops",
		Usage: "Simple CLI to deploy apps to VPS",
		Commands: []*cli.Command{
			cmd.ContextCommand,
			cmd.DeployCommand,
			cmd.ProxyCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
