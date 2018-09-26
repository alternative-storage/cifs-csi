package main

import (
	"flag"
	"os"

	"github.com/alternative-storage/cifs-csi/pkg/cifs"
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
	driver := cifs.NewCifsDriver()
	driver.Init(*driverName, *nodeId)
	driver.Start(*endpoint)
	os.Exit(0)
}
