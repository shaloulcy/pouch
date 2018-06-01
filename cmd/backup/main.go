package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/alibaba/pouch/daemon/mgr"

	"github.com/sirupsen/logrus"
)

func main() {
	// container-root-path flag
	containerRootPath := flag.String("container-root-path", "/var/lib/pouch/containers", "pouch containers root path")

	// parse the flags
	flag.Parse()

	err := filepath.Walk(*containerRootPath, func(cPath string, f os.FileInfo, err error) error {
		if f.IsDir() && cPath != *containerRootPath {
			containerMeta := new(mgr.Container)
			metaPath := path.Join(cPath, "meta.json")
			backPath := path.Join(cPath, "meta.json.bak")

			fi, err := os.Stat(metaPath)
			if err != nil {
				logrus.Warnf("ignore %s, stat error: %v", metaPath, err)
				return nil
			} else if !fi.Mode().IsRegular() {
				logrus.Warnf("ignore %s, it is not a regular file", metaPath)
				return nil
			}

			// read the container meta
			raw, err := ioutil.ReadFile(metaPath)
			if err != nil {
				logrus.Warnf("ignore %s, readfile error: %v", metaPath, err)
				return nil
			}

			// backup the meta.json
			err = ioutil.WriteFile(backPath, raw, 0644)
			if err != nil {
				logrus.Warnf("ignore %s, backup failed: %v", metaPath, err)
				return nil
			}

			// unmarshal the container meta
			err = json.Unmarshal(raw, containerMeta)
			if err != nil {
				logrus.Warnf("ignore %s, unmarshal failed: %v", metaPath, err)
				return nil
			}

			// set DisableNetworkFiles to True
			containerMeta.Config.DisableNetworkFiles = true

			updateRaw, err := json.Marshal(containerMeta)
			if err != nil {
				logrus.Warnf("ignore %s, marshal failed: %v", metaPath, err)
				return nil
			}

			err = ioutil.WriteFile(metaPath, updateRaw, 0644)
			if err != nil {
				logrus.Errorf("write meta %s failed: %v", metaPath, err)
			}

			logrus.Infof("update %s successfully", metaPath)
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("Walk error: %v", err)
	}
}
