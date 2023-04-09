package commands

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"verdaccio-patch/patcher"
	"verdaccio-patch/utils"
	"verdaccio-patch/verdaccio"
)

var AdjustCommand = &cli.Command{
	Name:  "adjust",
	Usage: "Adjust all packages of local verdaccio storage",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "storage",
			Required: false,
			Usage:    "specify path of local verdaccio storage",
		},
	},
	Action: func(context *cli.Context) error {
		storagePath := context.String("storage")
		if storagePath == "" {
			storagePath = verdaccio.GetVerdaccioStoragePath()
		}
		if !filepath.IsAbs(storagePath) {
			if pwd, err := os.Getwd(); err == nil {
				storagePath = filepath.Join(pwd, storagePath)
			} else {
				return err
			}
		}
		if !utils.IsExists(storagePath) {
			fmt.Printf("%s：目录不存在\n", storagePath)
			return nil
		}
		dirs, err := os.ReadDir(storagePath)

		bar := progressbar.Default(int64(len(dirs)), "正在整理storage...")
		for _, dir := range dirs {
			bar.Add(1)
			if !dir.IsDir() {
				continue
			}
			err := patcher.AdjustPackage(filepath.Join(storagePath, dir.Name()))
			if err != nil {
				fmt.Printf("adjust %s failed\n", dir.Name())
			}
		}
		return err
	},
}
