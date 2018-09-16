package cifs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume/util"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type nodeServer struct {
	userCr *credentials
	*csicommon.DefaultNodeServer
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	glog.Infof("stage")
	var (
		err error
	)

	ns.userCr, err = getUserCredentials(req.GetNodeStageSecrets())

	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials from node stage secrets: %v", err)
	}
	if ns.userCr.id == "" || ns.userCr.key == "" {
		return nil, fmt.Errorf("TODO: need to auth")
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.Infof("publish")
	targetPath := req.GetTargetPath()
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	mo := req.GetVolumeCapability().GetMount().GetMountFlags()
	if req.GetReadonly() {
		mo = append(mo, "ro")
	}

	mo = append(mo, fmt.Sprintf("username=%s", ns.userCr.id))
	mo = append(mo, fmt.Sprintf("password=%s", ns.userCr.key))

	s := req.GetVolumeAttributes()["server"]
	ep := req.GetVolumeAttributes()["share"]
	source := fmt.Sprintf("//%s/%s", s, ep)

	mounter := mount.New("")
	err = mounter.Mount(source, targetPath, "cifs", mo)
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

const (
	credUserId  = "userID"
	credUserKey = "userKey"
)

type credentials struct {
	id  string
	key string
}

func getCredentials(idField, keyField string, secrets map[string]string) (*credentials, error) {
	var (
		c  = &credentials{}
		ok bool
	)

	if c.id, ok = secrets[idField]; !ok {
		return nil, fmt.Errorf("missing ID field '%s' in secrets", idField)
	}

	if c.key, ok = secrets[keyField]; !ok {
		return nil, fmt.Errorf("missing key field '%s' in secrets", keyField)
	}

	return c, nil
}

func getUserCredentials(secrets map[string]string) (*credentials, error) {
	return getCredentials(credUserId, credUserKey, secrets)
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Error(codes.NotFound, "Targetpath not found")
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if notMnt {
		return nil, status.Error(codes.NotFound, "Volume not mounted")
	}

	err = util.UnmountPath(req.GetTargetPath(), mount.New(""))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}
