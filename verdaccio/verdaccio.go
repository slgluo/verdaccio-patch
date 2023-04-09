package verdaccio

import (
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

type VerdaccioConfig struct {
	Storage string `yaml:"storage"`
}

func GetVerdaccioHome() string {
	// windows
	userDataDir := os.Getenv("XDG_DATA_HOME")
	// linux or macos
	if userDataDir == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		userDataDir = filepath.Join(u.HomeDir, ".config")
	}
	verdaccioHome := filepath.Join(userDataDir, "verdaccio")
	return verdaccioHome
}

// h k s 小四号
func GetVerdaccioConfig() VerdaccioConfig {
	content, err := os.ReadFile(filepath.Join(GetVerdaccioHome(), "config.yaml"))
	if err != nil {
		panic(err)
	}

	config := VerdaccioConfig{}
	unmarshalErr := yaml.Unmarshal(content, &config)
	if unmarshalErr != nil {
		panic(unmarshalErr)
	}
	return config
}

func GetVerdaccioStoragePath() string {
	storagePath := os.Getenv("VERDACCIO_STORAGE_PATH")
	if storagePath == "" {
		storage := GetVerdaccioConfig().Storage
		if filepath.IsAbs(storage) {
			storagePath = storage
		} else {
			storagePath = filepath.Join(GetVerdaccioHome(), storage)
		}
	}
	return storagePath
}

func GeStoragePackages() []string {
	storagePath := GetVerdaccioStoragePath()
	dirs, err := os.ReadDir(storagePath)
	if err != nil {
		log.Fatal(err)
	}
	packages := lo.Map(dirs, func(item os.DirEntry, _ int) string {
		return item.Name()
	})
	return packages
}
