package cifs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/golang/glog"
)

const (
	controllerCacheRoot = PluginFolder + "/controller/plugin-cache"
)

type controllerCacheEntry struct {
	VolOptions volumeOptions
	VolumeID   volumeID
}

type controllerCacheMap map[volumeID]*controllerCacheEntry

var (
	ctrCache    = make(controllerCacheMap)
	ctrCacheMtx sync.Mutex
)

// Load all .json files from controllerCacheRoot into ctrCache
// Called from driver.go's Run()
func loadControllerCache() error {
	cacheDir, err := ioutil.ReadDir(controllerCacheRoot)
	if err != nil {
		return fmt.Errorf("cannot read controller cache from %s: %v", controllerCacheRoot, err)
	}

	ctrCacheMtx.Lock()
	defer ctrCacheMtx.Unlock()

	for _, fi := range cacheDir {
		if !strings.HasSuffix(fi.Name(), ".json") || !fi.Mode().IsRegular() {
			continue
		}

		f, err := os.Open(path.Join(controllerCacheRoot, fi.Name()))
		if err != nil {
			glog.Errorf("cifs: cloudn't read '%s' from controller cache: %v", fi.Name(), err)
			continue
		}

		d := json.NewDecoder(f)
		ent := &controllerCacheEntry{}

		if err = d.Decode(ent); err != nil {
			glog.Errorf("cifs: failed to parse '%s': %v", fi.Name(), err)
		} else {
			ctrCache[ent.VolumeID] = ent
		}

		f.Close()
	}

	return nil
}

func getControllerCacheEntryPath(volId volumeID) string {
	return path.Join(controllerCacheRoot, string(volId)+".json")
}

func (m controllerCacheMap) insert(ent *controllerCacheEntry) error {
	filePath := getControllerCacheEntryPath(ent.VolumeID)

	ctrCacheMtx.Lock()
	defer ctrCacheMtx.Unlock()

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("couldn't create cache entry file '%s': %v", filePath, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err = enc.Encode(ent); err != nil {
		return fmt.Errorf("failed to encode cache entry for volume %s: %v", ent.VolumeID, err)
	}

	m[ent.VolumeID] = ent

	return nil
}

func (m controllerCacheMap) pop(volId volumeID) (*controllerCacheEntry, error) {
	ctrCacheMtx.Lock()
	defer ctrCacheMtx.Unlock()

	ent, ok := m[volId]
	if !ok {
		return nil, fmt.Errorf("cache entry for volume %s does not exist", volId)
	}

	filePath := getControllerCacheEntryPath(volId)

	if err := os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("failed to remove cache entry file '%s': %v", filePath, err)
	}

	delete(m, volId)

	return ent, nil
}
