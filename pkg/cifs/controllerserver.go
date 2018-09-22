package cifs

import (
	"github.com/golang/glog"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	oneGB = 1073741824
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.validateCreateVolumeRequest(req); err != nil {
		glog.Errorf("CreateVolumeRequest validation failed: %v", err)
		return nil, err
	}

	volId := newVolumeID()

	// TODO: Setting quota and attributes

	sz := req.GetCapacityRange().GetRequiredBytes()
	if sz == 0 {
		sz = oneGB
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            string(volId),
			CapacityBytes: sz,
			Attributes:    req.GetParameters(),
		},
	}, nil

}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.validateDeleteVolumeRequest(req); err != nil {
		glog.Errorf("DeleteVolumeRequest validation failed: %v", err)
		return nil, err
	}

	// TODO

	return &csi.DeleteVolumeResponse{}, nil

}

func newVolumeID() volumeID {
	return volumeID("csi-cifs-" + uuid.NewUUID().String())
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	r := &csi.ValidateVolumeCapabilitiesResponse{
		Supported: true,
	}
	// CIFS doesn't support Block volume
	for _, cap := range req.VolumeCapabilities {
		if t := cap.GetBlock(); t != nil {
			r.Supported = false
			break
		}
		if t := cap.GetMount(); t != nil {
			// If a filesystem is given, it must be cifs
			fs := t.GetFsType()
			if fs != "" && fs != "cifs" {
				r.Supported = false
				break
			}
		}
	}
	return r, nil
}
