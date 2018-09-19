package cifs

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/csi-test/utils"
)

func TestNodeServer(t *testing.T) {

	// Setup simple driver
	d := NewCifsDriver()
	go d.Run(driverName, nodeId, tcp_ep)
	defer d.Stop()

	// Setup a connection to the driver
	conn, err := utils.Connect(tcp_addr)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	defer conn.Close()

	// Make a call
	c := csi.NewNodeClient(conn)
	r, err := c.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	// Verify
	nid := r.GetNodeId()
	if nid != nodeId {
		t.Errorf("Unknown driver name: %s\n", nid)
	}
}
