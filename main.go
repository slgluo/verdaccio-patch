package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"verdaccio-patch/commands"
)

func main() {
	app := &cli.App{
		Name:    "verdaccio-patch",
		Usage:   "cli for patch verdaccio storage",
		Version: "V1.0.0",
		Authors: []*cli.Author{
			{
				Name:  "sunlg",
				Email: "slgluo@qq.com",
			},
		},
		//Action: func(ctx *cli.Context) error {
		//	fmt.Println("cli for patch verdaccio storage")
		//	return nil
		//},
		Commands: []*cli.Command{
			commands.PatchCommand,
			commands.AdjustCommand,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
