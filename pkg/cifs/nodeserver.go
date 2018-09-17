package cifs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

type volumeID string

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	glog.Infof("stage")
	var err error

	if err = validateNodeStageVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Configuration

	stagingTargetPath := req.GetStagingTargetPath()
	volId := volumeID(req.GetVolumeId())
	glog.Infof("cifs: volume %s is trying to create and mount %s", volId, stagingTargetPath)

	notMnt, err := mount.New("").IsLikelyNotMountPoint(stagingTargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(stagingTargetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if !notMnt {
		glog.Infof("cifs: volume %s is already mounted to %s, skipping", volId, stagingTargetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	ns.userCr, err = getUserCredentials(req.GetNodeStageSecrets())

	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials from node stage secrets: %v", err)
	}
	if ns.userCr.id == "" || ns.userCr.key == "" {
		return nil, fmt.Errorf("TODO: need to auth")
	}

	mo := []string{}
	mo = append(mo, fmt.Sprintf("username=%s", ns.userCr.id))
	mo = append(mo, fmt.Sprintf("password=%s", ns.userCr.key))

	s := req.GetVolumeAttributes()["server"]
	ep := req.GetVolumeAttributes()["share"]
	source := fmt.Sprintf("//%s/%s", s, ep)

	mounter := mount.New("")
	err = mounter.Mount(source, stagingTargetPath, "cifs", mo)
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.Infof("publish")

	if err := validateNodePublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	targetPath := req.GetTargetPath()
	volId := req.GetVolumeId()

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
		glog.Infof("cifs: volume %s is already bind-mounted to %s", volId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	if err = bindMount(req.GetStagingTargetPath(), req.GetTargetPath(), req.GetReadonly()); err != nil {
		glog.Errorf("failed to bind-mount volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.Infof("cifs: successfuly bind-mounted volume %s to %s", volId, targetPath)

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
	if err := validateNodeUnpublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if err := validateNodeUnstageVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	stagingTargetPath := req.GetStagingTargetPath()
	// Unmount the volume
	if err := util.UnmountPath(stagingTargetPath, mount.New("")); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	os.Remove(stagingTargetPath)

	glog.Infof("cephfs: successfuly umounted volume %s from %s", req.GetVolumeId(), stagingTargetPath)

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func bindMount(from, to string, readOnly bool) error {
	if err := execCommandAndValidate("mount", "--bind", from, to); err != nil {
		return fmt.Errorf("failed to bind-mount %s to %s: %v", from, to, err)
	}

	if readOnly {
		if err := execCommandAndValidate("mount", "-o", "remount,ro,bind", to); err != nil {
			return fmt.Errorf("failed read-only remount of %s: %v", to, err)
		}
	}

	return nil
}

func execCommand(command string, args ...string) ([]byte, error) {
	glog.V(4).Infof("cifs: EXEC %s %s", command, args)

	cmd := exec.Command(command, args...)
	return cmd.CombinedOutput()
}

func execCommandAndValidate(program string, args ...string) error {
	out, err := execCommand(program, args...)
	if err != nil {
		return fmt.Errorf("cifs: %s failed with following error: %s\ncifs: %s output: %s", program, err, program, out)
	}

	return nil
}
