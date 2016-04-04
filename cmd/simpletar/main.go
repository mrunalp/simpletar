package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/mrunalp/simpletar"
)

func main() {
	app := cli.NewApp()
	app.Name = "simpletar"
	app.Version = "0.1"
	app.Usage = "simple tar/untar utility"
	app.Commands = []cli.Command{
		tarCommand,
		untarCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var tarCommand = cli.Command{
	Name:  "tar",
	Usage: "tar a directory",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "source", Usage: "source directory path"},
		cli.StringFlag{Name: "dest", Usage: "destination tar file path"},
	},
	Action: func(ctx *cli.Context) {
		source := ctx.String("source")
		dest := ctx.String("dest")
		if source == "" {
			log.Fatalf("--source could not be emptry")
		}
		if dest == "" {
			log.Fatalf("--dest could not be empty")
		}
		if err := simpletar.Tar(source, dest); err != nil {
			log.Fatalf("Failed to tar: %v", err)
		}
	},
}

var untarCommand = cli.Command{
	Name:  "untar",
	Usage: "untar a file",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "source", Usage: "source tar file path"},
		cli.StringFlag{Name: "dest", Usage: "destination directory path"},
	},
	Action: func(ctx *cli.Context) {
		source := ctx.String("source")
		dest := ctx.String("dest")
		if source == "" {
			log.Fatalf("--source could not be emptry")
		}
		if dest == "" {
			log.Fatalf("--dest could not be empty")
		}
		if err := simpletar.Untar(source, dest); err != nil {
			log.Fatalf("Failed to tar: %v", err)
		}
	},
}
