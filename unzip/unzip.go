package unzip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Unzip(path, target string) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		fmt.Println(fmt.Errorf(err.Error()))
		return
	}
	defer archive.Close()
	fmt.Printf("unzipping file %s\n", path)

	for _, f := range archive.File {
		filePath := filepath.Join(target, f.Name)
		//fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(target)+string(os.PathSeparator)) {
			//fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			//fmt.Println("creating directory...")
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			fmt.Println(fmt.Errorf(err.Error()))
			return
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			fmt.Println(fmt.Errorf(err.Error()))
			return
		}

		fileInArchive, err := f.Open()
		if err != nil {
			fmt.Println(fmt.Errorf(err.Error()))
			return
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			fmt.Println(fmt.Errorf(err.Error()))
			return
		}

		dstFile.Close()
		fileInArchive.Close()
	}

}
