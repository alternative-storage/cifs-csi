package cifs

import (
	"context"
	"os"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/csi-test/utils"
	"k8s.io/kubernetes/pkg/util/mount"
)

func TestNodePublishVolume(t *testing.T) {
	// Setup simple driver
	d := NewCifsDriver()
	d.Init(driverName, nodeId)

	d.ns.mounter = &mount.FakeMounter{}
	go d.Start(tcp_ep)
	defer d.Stop()

	// Setup a connection to the driver
	conn, err := utils.Connect(tcp_addr)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	defer conn.Close()

	tests := []struct {
		name   string
		req    *csi.NodePublishVolumeRequest
		errors bool
	}{
		{
			name: "Success",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:           "testvol",
				StagingTargetPath:  "/tmp/stg",
				TargetPath:         "/tmp/tgt",
				Readonly:           false,
				NodePublishSecrets: map[string]string{"username": "user", "password": "pass"},
				VolumeAttributes:   map[string]string{"server": "example.com", "share": "test"},
			},
			errors: false,
		},
		{
			name: "Fail due to missing volume ID",
			req: &csi.NodePublishVolumeRequest{
				StagingTargetPath:  "/tmp/stg",
				TargetPath:         "/tmp/tgt",
				Readonly:           false,
				NodePublishSecrets: map[string]string{"username": "user", "password": "pass"},
				VolumeAttributes:   map[string]string{"server": "example.com", "share": "test"},
			},
			errors: true,
		},
		{
			name: "Fail due to missing target path",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:           "testvol",
				StagingTargetPath:  "/tmp/stg",
				Readonly:           false,
				NodePublishSecrets: map[string]string{"username": "user", "password": "pass"},
				VolumeAttributes:   map[string]string{"server": "example.com", "share": "test"},
			},
			errors: true,
		},
	}

	// Make a call
	c := csi.NewNodeClient(conn)

	for _, tc := range tests {
		_, err = c.NodePublishVolume(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpected error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
		d.ns.mounter.Unmount(tc.req.TargetPath)
	}
}

func TestNodeUnpublishVolume(t *testing.T) {
	// Setup simple driver
	d := NewCifsDriver()
	d.Init(driverName, nodeId)

	mp := mount.MountPoint{Device: "/dev/foo", Path: "/tmp/tgt"}
	d.ns.mounter = &mount.FakeMounter{MountPoints: []mount.MountPoint{mp}}
	go d.Start(tcp_ep)
	defer d.Stop()

	// Setup a connection to the driver
	conn, err := utils.Connect(tcp_addr)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	defer conn.Close()

	tests := []struct {
		name   string
		req    *csi.NodeUnpublishVolumeRequest
		errors bool
	}{
		{
			name: "Success",
			req: &csi.NodeUnpublishVolumeRequest{
				VolumeId:   "testvol",
				TargetPath: "/tmp/tgt",
			},
			errors: false,
		},
		{
			name: "Fail due to not mounted targetpath",
			req: &csi.NodeUnpublishVolumeRequest{
				VolumeId:   "testvol",
				TargetPath: "/tmp/wrong",
			},
			errors: true,
		},
	}

	// Make a call
	c := csi.NewNodeClient(conn)

	for _, tc := range tests {
		os.MkdirAll(tc.req.TargetPath, 0750)
		_, err = c.NodeUnpublishVolume(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpected error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
		d.ns.mounter.Unmount(tc.req.TargetPath)
	}
}
