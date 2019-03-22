package main

import (
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		wd = ""
	}

	app := cli.NewApp()
	app.Name = "crccheck"
	app.Usage = "extract the CRC value from file names and validate the files' integrity."
	app.Version = "1.0.0"
	app.Copyright = "(c) 2019 Dominik Nakamura"
	app.Authors = []cli.Author{
		{Name: "Dominik Nakamura", Email: "dnaka91@gmail.com"},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "dir, d",
			Usage: "directory to scan for files. If not set, defaults to the current working directory",
			Value: wd,
		},
		cli.BoolFlag{
			Name:  "update, u",
			Usage: "if the hash of a file mismatches, rename it to the new hash",
		},
	}
	app.Action = func(c *cli.Context) error {
		return check(afero.NewOsFs(), c.String("dir"), c.Bool("update"))
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
