package cifs

import (
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type identityServer struct {
	*csicommon.DefaultNodeServer
}
