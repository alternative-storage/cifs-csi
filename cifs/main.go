package main

import (
	"flag"
	"os"
	"path"

	"github.com/alternative-storage/cifs-csi/pkg/cifs"
	"github.com/golang/glog"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName = flag.String("drivername", "csi-cifsplugin", "name of the driver")
	nodeId     = flag.String("nodeid", "", "node id")
)

func main() {
	flag.Parse()

	if err := createPersistentStorage(path.Join(cifs.PluginFolder, "controller")); err != nil {
		glog.Errorf("failed to create persistent storage for controller: %v", err)
		os.Exit(1)
	}

	if err := createPersistentStorage(path.Join(cifs.PluginFolder, "node")); err != nil {
		glog.Errorf("failed to create persistent storage for node: %v", err)
		os.Exit(1)
	}

	driver := cifs.NewCifsDriver()
	driver.Init(*driverName, *nodeId)
	driver.Start(*endpoint)

	os.Exit(0)
}

func createPersistentStorage(persistentStoragePath string) error {
	return os.MkdirAll(persistentStoragePath, os.FileMode(0755))
}
