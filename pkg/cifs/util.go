package cifs

import (
	"fmt"
	"os/exec"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
)

func (cs *controllerServer) validateCreateVolumeRequest(req *csi.CreateVolumeRequest) error {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return fmt.Errorf("invalid CreateVolumeRequest: %v", err)
	}

	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "Volume Name cannot be empty")
	}

	return nil
}

func (cs *controllerServer) validateDeleteVolumeRequest(req *csi.DeleteVolumeRequest) error {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return fmt.Errorf("invalid DeleteVolumeRequest: %v", err)
	}

	return nil
}

func validateNodeStageVolumeRequest(req *csi.NodeStageVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("volume ID missing in request")
	}

	if req.GetStagingTargetPath() == "" {
		return fmt.Errorf("staging target path missing in request")
	}

	if req.GetNodeStageSecrets() == nil || len(req.GetNodeStageSecrets()) == 0 {
		return fmt.Errorf("stage secrets cannot be nil or empty")
	}

	return nil
}

func validateNodePublishVolumeRequest(req *csi.NodePublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("volume ID missing in request")
	}

	if req.GetStagingTargetPath() == "" {
		return fmt.Errorf("staging target path missing in request")
	}

	if req.GetTargetPath() == "" {
		return fmt.Errorf("varget path missing in request")
	}

	return nil
}

func validateNodeUnpublishVolumeRequest(req *csi.NodeUnpublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return fmt.Errorf("target path missing in request")
	}

	return nil
}

func validateNodeUnstageVolumeRequest(req *csi.NodeUnstageVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return fmt.Errorf("volume ID missing in request")
	}

	if req.GetStagingTargetPath() == "" {
		return fmt.Errorf("staging target path missing in request")
	}
	return nil
}

type Interface interface {
	execCommand() ([]byte, error)
	execCommandAndValidate() error
}

var _ Interface = &fakeCommander{}
var _ Interface = &commander{}

type commander struct {
	cmd     string
	options []string
}

type fakeCommander struct {
	commander
}

func (c *commander) execCommand() ([]byte, error) {
	glog.V(4).Infof("cifs: EXEC %s %s", c.cmd, c.options)

	cmd := exec.Command(c.cmd, c.options...)
	return cmd.CombinedOutput()
}

func (c *commander) execCommandAndValidate() error {
	out, err := c.execCommand()
	if err != nil {
		return fmt.Errorf("cifs: %s failed with following error: %s\ncifs: %s output: %s", c.cmd, err, c.cmd, out)
	}

	return nil
}

func (c *fakeCommander) execCommandAndValidate() error {
	return nil
}

func (c *fakeCommander) execCommand() ([]byte, error) {
	return nil, nil
}
