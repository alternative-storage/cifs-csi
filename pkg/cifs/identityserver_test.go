package cifs

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/csi-test/utils"
)

func TestIdentityServer(t *testing.T) {

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

	// Make a call
	c := csi.NewIdentityClient(conn)
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	// Verify
	name := r.GetName()
	if name != driverName {
		t.Errorf("Unknown driver name: %s\n", name)
	}

	ver := r.GetVendorVersion()
	if ver != Version {
		t.Errorf("Unknown driver version: %s\n", ver)
	}
}
