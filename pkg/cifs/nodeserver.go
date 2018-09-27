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
	cr *credentials
	*csicommon.DefaultNodeServer

	mounter mount.Interface
}

type volumeID string

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.Infof("publish")

	if err := validateNodePublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if ns.mounter == nil {
		ns.mounter = mount.New("")
	}

	targetPath := req.GetTargetPath()
	volId := req.GetVolumeId()

	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)
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
		glog.Infof("cifs: volume %s is already bind-mounted to %s", volId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	ns.cr, err = getUserCredentials(req.GetNodePublishSecrets())

	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials from node stage secrets: %v", err)
	}
	if ns.cr.username == "" || ns.cr.password == "" {
		return nil, fmt.Errorf("TODO: need to auth")
	}

	mo := []string{}
	mo = append(mo, fmt.Sprintf("username=%s", ns.cr.username))
	mo = append(mo, fmt.Sprintf("password=%s", ns.cr.password))

	s := req.GetVolumeAttributes()["server"]
	if s == "" {
		return nil, fmt.Errorf("TODO: need server or endpoint")
	}
	source := fmt.Sprintf("//%s/%s", s, volId)

	err = ns.mounter.Mount(source, targetPath, "cifs", mo)
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.Infof("cifs: successfully mounted volume %s to %s", volId, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

const (
	username = "username"
	password = "password"

	admin_name     = "admin_name"
	admin_password = "admin_password"
)

type credentials struct {
	username string
	password string
}

func getCredentials(u, p string, secrets map[string]string) (*credentials, error) {
	var (
		c  = &credentials{}
		ok bool
	)

	if c.username, ok = secrets[u]; !ok {
		return nil, fmt.Errorf("missing username in secrets")
	}

	if c.password, ok = secrets[p]; !ok {
		return nil, fmt.Errorf("missing password in secrets")
	}

	return c, nil
}

func getUserCredentials(secrets map[string]string) (*credentials, error) {
	return getCredentials(username, password, secrets)
}

func getAdminCredentials(secrets map[string]string) (*credentials, error) {
	return getCredentials(admin_name, admin_password, secrets)
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if err := validateNodeUnpublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if ns.mounter == nil {
		ns.mounter = mount.New("")
	}

	targetPath := req.GetTargetPath()
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)

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
