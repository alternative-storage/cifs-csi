package cifs

const (
	driverName = "mock"
	nodeId     = "cifs-mock-node"

	tcp_schema = "tcp://"
	tcp_addr   = "127.0.0.1:10000"
	tcp_ep     = tcp_schema + tcp_addr

	unix_schema = "unix://"
	unix_sock   = "/tmp/foo.sock"
	unix_ep     = unix_schema + unix_sock
)
