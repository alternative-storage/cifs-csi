package cifs

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	oneGB = 1073741824
)

type controllerServer struct {
	cr *credentials
	*csicommon.DefaultControllerServer

	commander Interface
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.validateCreateVolumeRequest(req); err != nil {
		glog.Errorf("CreateVolumeRequest validation failed: %v", err)
		return nil, err
	}

	volOptions, err := newVolumeOptions(req.GetParameters())
	if err != nil {
		glog.Errorf("validation of volume options failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	volId := newVolumeID()
	sharep := string(volId) + "=" + req.GetParameters()["path"]
	volOptions.Share = string(volId)

	cs.cr, err = getAdminCredentials(req.GetControllerCreateSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to get admin credentials from create volume secrets: %v", err)
	}
	adminpass := fmt.Sprintf("%s%%%s", cs.cr.username, cs.cr.password)
	// TODO port?
	debug := "1"
	if glog.V(4) {
		debug = "4"
	}
	if cs.commander == nil {
		// $ net rpc share add SHARE_NAME=/PATH/TO/SHARE COMMENT -S server -d 4
		args := []string{
			"rpc", "share", "add", sharep, "\"test comment\"",
			"-S", volOptions.Server, "-d", debug, "-U", adminpass,
		}
		cs.commander = &commander{cmd: "net", options: args}
	}
	if err := cs.commander.execCommandAndValidate(); err != nil {
		return nil, err
	}
	defer func() { cs.commander = nil }()

	// TODO: Setting quota and attributes

	sz := req.GetCapacityRange().GetRequiredBytes()
	if sz == 0 {
		sz = oneGB
	}

	if err = ctrCache.insert(&controllerCacheEntry{VolOptions: *volOptions, VolumeID: volId}); err != nil {
		glog.Errorf("failed to store a cache entry for volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
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

	var (
		volId = volumeID(req.GetVolumeId())
		err   error
	)

	// Load volume info from cache
	ent, err := ctrCache.pop(volId)
	if err != nil {
		glog.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer func() {
		if err != nil {
			// Reinsert cache entry for retry
			if insErr := ctrCache.insert(ent); insErr != nil {
				glog.Errorf("failed to reinsert volume cache entry in rollback procedure for volume %s: %v", volId, err)
			}
		}
	}()

	cs.cr, err = getAdminCredentials(req.GetControllerDeleteSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to get admin credentials from create volume secrets: %v", err)
	}
	adminpass := fmt.Sprintf("%s%%%s", cs.cr.username, cs.cr.password)

	// TODO port?

	debug := "1"
	if glog.V(5) {
		debug = "4"
	}
	if cs.commander == nil {
		// $ net rpc share delete $SHARE -S $SERVER -U root%xxx
		args := []string{
			"rpc", "share", "delete", ent.VolOptions.Share,
			"-S", ent.VolOptions.Server, "-d", debug, "-U", adminpass,
		}
		cs.commander = &commander{cmd: "net", options: args}
	}
	defer func() { cs.commander = nil }()
	if err := cs.commander.execCommandAndValidate(); err != nil {
		return nil, err
	}

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
