package patcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/tidwall/pretty"
	"os"
	"path/filepath"
	"strings"
	"verdaccio-patch/dependency"
	"verdaccio-patch/utils"
	"verdaccio-patch/verdaccio"
)

func PatchStorage(src string) {
	needPatchPackages, err := os.ReadDir(src)
	if err != nil {
		fmt.Printf("读取patch storage目录失败：%s\n", src)
		return
	}

	storagePath := verdaccio.GetVerdaccioStoragePath()

	for _, pkg := range needPatchPackages {
		srcPkg, targetPkg := filepath.Join(src, pkg.Name()), filepath.Join(storagePath, pkg.Name())
		err := PatchPackage(srcPkg, targetPkg)
		if err != nil {
			fmt.Printf("patch [%s] 失败\n", pkg.Name())
		} else {
			fmt.Printf("%s patched\n", pkg.Name())
		}
	}
}

func AdjustPackage(packagePath string) error {
	var (
		pkg   *dependency.Package
		dists []string
		err   error
	)
	jsonPath := filepath.Join(packagePath, "package.json")
	pkg, err = dependency.GetPackage(jsonPath)
	if err != nil {
		dirs, err := os.ReadDir(packagePath)
		if err == nil && len(dirs) > 0 {
			for _, dir := range dirs {
				srcPkg := filepath.Join(packagePath, dir.Name())
				err = AdjustPackage(srcPkg)
				if err != nil {
					fmt.Printf("patch package [%s] 失败\n", srcPkg)
				}
			}
			return nil
		} else {
			fmt.Printf("读取package.json失败：%s\n", jsonPath)
			return errors.Unwrap(err)
		}
	}
	dists, err = dependency.GetLocalDistFiles(packagePath)
	if err != nil {
		fmt.Printf("获取获取本地依赖包发布版失败：%s\n", packagePath)
		return errors.Unwrap(err)
	}
	localVersions := dependency.GetVersions(dists)
	// 更新versions字段
	newVersions := make(map[string]interface{})
	for k, v := range pkg.Versions {
		if lo.Contains(localVersions, k) {
			newVersions[k] = v
		}
	}
	pkg.Versions = newVersions

	// 更新time字段
	newTime := make(map[string]string)
	for k, v := range pkg.Time {
		if lo.Contains(localVersions, k) {
			newTime[k] = v
		}
	}
	versionsInTime := dependency.GetSortedVersions(lo.Keys(newTime))
	timeLen := len(versionsInTime)
	if timeLen > 0 {
		newTime["created"] = newTime[versionsInTime[timeLen-1]]
		newTime["modified"] = newTime[versionsInTime[0]]
	}
	pkg.Time = newTime

	// 更新_distfiles字段
	newDistFiles := make(map[string]dependency.DistFile)
	for k, v := range pkg.DistFiles {
		if lo.Contains(dists, k) {
			newDistFiles[k] = v
		}
	}
	pkg.DistFiles = newDistFiles

	// 更新_attachment字段
	newAttachments := make(map[string]dependency.Attachment)
	for k, v := range pkg.Attachments {
		if lo.Contains(dists, k) {
			newAttachments[k] = v
		}
	}
	pkg.Attachments = newAttachments

	// 更新 dist-tags字段（由于无法从版本号中判断出除latest之外的标签，其他标签的更新会有问题）
	newDistTags := make(map[string]string)
	for k, v := range pkg.DistTags {
		if k == "latest" {
			newDistTags[k] = dependency.GetVersionFromDistFile(dependency.GetLatestDist(dists))
		} else if lo.Contains(localVersions, v) {
			newDistTags[k] = v
		}
	}
	pkg.DistTags = newDistTags

	content, err := json.Marshal(pkg)
	content = pretty.Pretty(content)

	err = os.WriteFile(jsonPath, content, 0777)

	return err
}

func mergePackageJson(src, dest string) (*dependency.Package, error) {
	var (
		srcPkg  *dependency.Package
		destPkg *dependency.Package
		err     error
	)
	srcPkg, err = dependency.GetPackage(src)
	if err != nil {
		fmt.Printf("读取 package.json 失败：%s\n", src)
		return nil, err
	}

	destPkg, err = dependency.GetPackage(dest)
	if err != nil {
		fmt.Printf("读取 package.json 失败：%s\n", dest)
		return nil, err
	}
	// 合并 versiongs
	for k, v := range srcPkg.Versions {
		destPkg.Versions[k] = v
	}

	// 合并 time
	for k, v := range srcPkg.Time {
		destPkg.Time[k] = v
	}

	// 合并 dist-tags
	for k, v := range srcPkg.DistTags {
		destPkg.DistTags[k] = v
	}

	// 合并 _distfiles
	for k, v := range srcPkg.DistFiles {
		destPkg.DistFiles[k] = v
	}

	// 合并 _attachments
	for k, v := range srcPkg.Attachments {
		destPkg.Attachments[k] = v
	}

	return destPkg, err
}

func PatchPackage(srcPkgPath, targetPkgPath string) error {
	var err error
	if strings.HasPrefix(filepath.Base(srcPkgPath), "@") {
		subPackages, err := os.ReadDir(srcPkgPath)
		if err != nil {
			fmt.Printf("读取目录失败：%s\n", srcPkgPath)
			return err
		}
		for _, pkg := range subPackages {
			a, b := filepath.Join(srcPkgPath, pkg.Name()), filepath.Join(targetPkgPath, pkg.Name())
			err := PatchPackage(a, b)
			if err != nil {
				fmt.Printf("patch [%s/%s] 失败\n", filepath.Base(a), pkg)
			}
		}
		return nil
	}

	srcPkgJson, destPkgJson := filepath.Join(srcPkgPath, "package.json"), filepath.Join(targetPkgPath, "package.json")
	// 如果已经存在
	if utils.IsExists(targetPkgPath) {
		// 合并package.json
		pkg, err := mergePackageJson(srcPkgJson, destPkgJson)
		if err != nil {
			fmt.Printf("合并 package.json失败: [%s -> %s]\n", srcPkgJson, destPkgJson)
			return err
		}
		content, err := json.Marshal(pkg)
		content = pretty.Pretty(content)
		err = os.WriteFile(destPkgJson, content, 0777)

		// 获取依赖包的所有版本
		versions, err := os.ReadDir(srcPkgPath)
		if err != nil {
			fmt.Printf("读取目录失败：%s\n", srcPkgPath)
			return err
		}
		// 遍历依赖包的所有版本
		for _, version := range versions {
			if version.Name() == "package.json" {
				continue
			}
			// 判断该版本是否已存在
			src := filepath.Join(srcPkgPath, version.Name())
			dest := filepath.Join(targetPkgPath, version.Name())
			if utils.IsExists(dest) {
				continue
			}
			// 如果不存在就复制到目标目录
			err = utils.Copy(src, dest)

			if err != nil {
				fmt.Printf("复制 %s 失败\n", version.Name())
				return err
			}
		}

		// 整理package.json
		err = AdjustPackage(targetPkgPath)
		if err != nil {
			fmt.Printf("整理 package.json 失败：%s\n", filepath.Join(targetPkgPath, "package.json"))
			return err
		}
	} else {
		// 如果不存在，拷贝到目标目录
		err = utils.Copy(srcPkgPath, targetPkgPath)
		if err != nil {
			fmt.Printf("复制 %s 失败\n", filepath.Base(srcPkgPath))
			return err
		}

		err = AdjustPackage(targetPkgPath)
		if err != nil {
			fmt.Printf("整理 package.json 失败：%s\n", filepath.Join(targetPkgPath, "package.json"))
			return err
		}
	}
	return err
}
