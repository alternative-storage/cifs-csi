package cifs

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/csi-test/utils"
	"k8s.io/kubernetes/pkg/util/mount"
)

func TestNodeStageVolume(t *testing.T) {
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
		req    *csi.NodeStageVolumeRequest
		errors bool
	}{
		{
			name: "Success",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "testvol",
				StagingTargetPath: "/tmp/foo",
				NodeStageSecrets:  map[string]string{"userID": "user", "userKey": "pass"},
				VolumeAttributes:  map[string]string{"server": "example.com", "share": "test"},
			},
			errors: false,
		},
		{
			name: "Fail due to missing volume",
			req: &csi.NodeStageVolumeRequest{
				StagingTargetPath: "/tmp/foo",
				NodeStageSecrets:  map[string]string{"userID": "user", "userKey": "pass"},
				VolumeAttributes:  map[string]string{"server": "example.com", "share": "test"},
			},
			errors: true,
		},
		{
			name: "Fail due to missing secrets",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "testvol",
				StagingTargetPath: "/tmp/foo",
				NodeStageSecrets:  map[string]string{"userKey": "pass"},
				VolumeAttributes:  map[string]string{"server": "example.com", "share": "test"},
			},
			errors: true,
		},
		{
			name: "Fail due to missing targetpath",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:         "testvol",
				NodeStageSecrets: map[string]string{"userID": "user", "userKey": "pass"},
				VolumeAttributes: map[string]string{"server": "example.com", "share": "test"},
			},
			errors: true,
		},
		{
			name: "Fail due to missing server name",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "testvol",
				StagingTargetPath: "/tmp/foo",
				NodeStageSecrets:  map[string]string{"userID": "user", "userKey": "pass"},
				VolumeAttributes:  map[string]string{"share": "test"},
			},
			errors: true,
		},
	}

	// Make a call
	c := csi.NewNodeClient(conn)

	for _, tc := range tests {
		_, err = c.NodeStageVolume(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpectd error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
	}
}

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
				VolumeId:          "testvol",
				StagingTargetPath: "/tmp/stg",
				TargetPath:        "/tmp/tgt",
				Readonly:          false,
			},
			errors: false,
		},
		{
			name: "Fail due to missing volume ID",
			req: &csi.NodePublishVolumeRequest{
				StagingTargetPath: "/tmp/stg",
				TargetPath:        "/tmp/tgt",
				Readonly:          false,
			},
			errors: true,
		},
		{
			name: "Fail due to missing staging target path",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   "testvol",
				TargetPath: "/tmp/tgt",
				Readonly:   false,
			},
			errors: true,
		},
		{
			name: "Fail due to missing target path",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:          "testvol",
				StagingTargetPath: "/tmp/stg",
				Readonly:          false,
			},
			errors: true,
		},
	}

	// Make a call
	c := csi.NewNodeClient(conn)

	for _, tc := range tests {
		_, err = c.NodePublishVolume(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpectd error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
	}
}
