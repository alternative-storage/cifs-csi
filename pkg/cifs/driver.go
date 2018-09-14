package cifs

import (
	"github.com/golang/glog"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	PluginFolder = "/var/lib/kubelet/plugins/csi-cifsplugin"
	Version      = "0.3.0"
)

type cifsDriver struct {
	driver *csicommon.CSIDriver

	is *identityServer
	ns *nodeServer
	cs *controllerServer

	caps   []*csi.VolumeCapability_AccessMode
	cscaps []*csi.ControllerServiceCapability
}

func NewCifsDriver() *cifsDriver {
	return &cifsDriver{}
}

func (fs *cifsDriver) Run(driverName, nodeId, endpoint, volumeMounter string) {
	glog.Infof("Driver: %v version: %v", driverName, Version)

}
