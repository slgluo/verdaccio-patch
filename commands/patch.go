package commands

import (
	"errors"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"verdaccio-patch/patcher"
	"verdaccio-patch/unzip"
	"verdaccio-patch/utils"
	"verdaccio-patch/verdaccio"
)

var PatchCommand = &cli.Command{
	Name:      "patch",
	Usage:     "Patch a extra storage to local verdaccio storage",
	ArgsUsage: "/path/to/storage-patch.zip",
	Action: func(context *cli.Context) error {
		patchPath := context.Args().Get(0)
		if patchPath == "" {
			fmt.Printf("Usage: %s %s /path/to/storage-patch.zip\n", context.App.Name, context.Command.Name)
			return nil
		}

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		patchPath = filepath.Clean(patchPath)

		if !filepath.IsAbs(patchPath) {
			patchPath = filepath.Join(pwd, patchPath)
		}

		if !utils.IsExists(patchPath) {
			return errors.New(fmt.Sprintf("%s 不存在", patchPath))
		}

		unzip.Unzip(patchPath, pwd)

		patchDir := filepath.Join(pwd, "storage-patch")

		needPatchPackages, err := os.ReadDir(patchDir)
		if err != nil {
			return errors.New(fmt.Sprintf("读取patch storage目录失败：%s\n", patchDir))
		}

		storagePath := verdaccio.GetVerdaccioStoragePath()

		bar := progressbar.Default(int64(len(needPatchPackages)), "patching...")
		for _, pkg := range needPatchPackages {
			bar.Add(1)
			srcPkg, targetPkg := filepath.Join(patchDir, pkg.Name()), filepath.Join(storagePath, pkg.Name())
			err := patcher.PatchPackage(srcPkg, targetPkg)
			if err != nil {
				fmt.Printf("patch [%s] 失败\n", pkg.Name())
			}
		}

		return os.RemoveAll(patchDir)
	},
}
