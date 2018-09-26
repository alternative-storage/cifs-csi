package cifs

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/csi-test/utils"
)

func TestCreateVolume(t *testing.T) {
	// Setup simple driver
	d := NewCifsDriver()
	d.Init(driverName, nodeId)
	d.cs.commander = &fakeCommander{}

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
		req    *csi.CreateVolumeRequest
		errors bool
		expId  string
	}{
		{
			name: "Success",
			req: &csi.CreateVolumeRequest{
				ControllerCreateSecrets: map[string]string{"admin_name": "user", "admin_password": "pass"},
				Parameters:              map[string]string{"server": "192.168.122.1"},
				Name:                    "testvol",
			},
			errors: false,
			expId:  "foo",
		},
		{
			name: "Fail due to missing password",
			req: &csi.CreateVolumeRequest{
				ControllerCreateSecrets: map[string]string{"admin_name": "user"},
				Name: "testvol",
			},
			errors: true,
		},
	}

	// Make a call
	c := csi.NewControllerClient(conn)
	for _, tc := range tests {
		res, err := c.CreateVolume(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpected error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
		if err == nil && res.Volume.Id == "" {
			t.Errorf("%s: expected volume ID", tc.name)
		}
	}
}

const testVID = "csi-cifs-testvol"

func TestDeleteVolume(t *testing.T) {
	// Setup simple driver
	d := NewCifsDriver()
	d.Init(driverName, nodeId)
	d.cs.commander = &fakeCommander{}

	go d.Start(tcp_ep)
	defer d.Stop()

	// Setup a connection to the driver
	conn, err := utils.Connect(tcp_addr)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	defer conn.Close()

	volOptions := &volumeOptions{Server: "192.168.122.1", Share: "testshare"}

	if err = ctrCache.insert(&controllerCacheEntry{VolOptions: *volOptions, VolumeID: volumeID(testVID)}); err != nil {
		t.Errorf("failed to store a cache entry for volume %s: %v", testVID, err)
	}

	tests := []struct {
		name   string
		req    *csi.DeleteVolumeRequest
		errors bool
		expId  string
	}{
		{
			name: "Success",
			req: &csi.DeleteVolumeRequest{
				VolumeId:                testVID,
				ControllerDeleteSecrets: map[string]string{"admin_name": "user", "admin_password": "pass"},
			},
			errors: false,
		},
	}

	// Make a call
	c := csi.NewControllerClient(conn)
	for _, tc := range tests {
		_, err := c.DeleteVolume(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpected error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
	}
}

func TestValidateVolumeCapabilities(t *testing.T) {
	// Setup simple driver
	d := NewCifsDriver()
	d.Init(driverName, nodeId)

	go d.Start(tcp_ep)
	defer d.Stop()

	// Setup a connection to the driver
	conn, err := utils.Connect(tcp_addr)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	defer conn.Close()

	tests := []struct {
		name       string
		req        *csi.ValidateVolumeCapabilitiesRequest
		errors     bool
		expSupport bool
	}{
		{
			name: "Success",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId:           "testvol",
				VolumeCapabilities: []*csi.VolumeCapability{},
			},
			errors:     false,
			expSupport: true,
		},
		{
			name: "Not supported as Block",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "testvol",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Block{
							Block: &csi.VolumeCapability_BlockVolume{},
						},
					},
				},
			},
			errors:     false,
			expSupport: false,
		},
		{
			name: "Supported as MOUNT",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "testvol",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
					},
				},
			},
			errors:     false,
			expSupport: true,
		},
		{
			name: "Support as cifs",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "testvol",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{FsType: "cifs"},
						},
					},
				},
			},
			errors:     false,
			expSupport: true,
		},
		{
			name: "Not support as non-cifs",
			req: &csi.ValidateVolumeCapabilitiesRequest{
				VolumeId: "testvol",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{FsType: "nfs"},
						},
					},
				},
			},
			errors:     false,
			expSupport: false,
		},
	}

	// Make a call
	c := csi.NewControllerClient(conn)
	for _, tc := range tests {
		res, err := c.ValidateVolumeCapabilities(context.Background(), tc.req)
		if err != nil && !tc.errors {
			t.Errorf("%s: unexpected error %v", tc.name, err.Error())
		}
		if err == nil && tc.errors {
			t.Errorf("%s: expected error, but not got any error", tc.name)
		}
		if res.Supported != tc.expSupport {
			t.Errorf("%s: expected supported as %v, but got %v", tc.name, tc.expSupport, res.Supported)
		}
	}
}
