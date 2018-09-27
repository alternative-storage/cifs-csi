[![Build Status](https://travis-ci.org/alternative-storage/cifs-csi.svg?branch=master)](https://travis-ci.org/alternative-storage/cifs-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/alternative-storage/cifs-csi)](https://goreportcard.com/report/github.com/alternative-storage/cifs-csi)

# CSI driver for CIFS

[Container Storage Interface (CSI)](https://github.com/container-storage-interface/) driver, provisioner, and attacher for CIFS (SMB, Samba, Windows Share) network filesystems.

## Supported matrix

Client         | Target         | Status         |
-------------- | -------------- | -------------- |
Linux          | Linux(Samba)   | WIP            |
Linux          | Windows        | -              |
Windows        | Linux(Samba)   | -              |
Windows        | Windows        | -              |


## Test

NOTE: First, you must change your samba server to accept `net rpc {add,delete}`. Please refer to [example steps](https://github.com/alternative-storage/cifs-csi/blob/master/examples/samba/README.md)

Get ```csc``` tool from https://github.com/rexray/gocsi/tree/master/csc


#### Get plugin info
```
$ csc identity plugin-info --endpoint tcp://127.0.0.1:10000
"csi-cifsplugin"	"0.3.0"
```

#### Create a volume
```
$ export X_CSI_SECRETS=admin_name="YOUR CIFS ADMIN USER",admin_password="YOUR CIFS ADMIN PASSWORD"

$ csc controller --endpoint tcp://127.0.0.1:10000 create-volume \
                 --params server=$CIFS_SERVER --params path="/tmp" \
                 testvol
csi-cifs-9bd0415d-c226-11e8-8086-54e1ad486e52
```

#### NodePublish a volume

```
$ export CIFS_SERVER="Your Server IP (Ex: 10.10.10.10)"
$ csc node publish --endpoint tcp://127.0.0.1:10000 \
                 --target-path /mnt/cifs \
                 --attrib server=192.168.121.127  csi-cifs-9bd0415d-c226-11e8-8086-54e1ad486e52
csi-cifs-9bd0415d-c226-11e8-8086-54e1ad486e52
```

#### NodeUnpublish a volume
```
$ csc node unpublish --endpoint tcp://127.0.0.1:10000 \
                    --target-path /mnt/cifs \
                    csi-cifs-9bd0415d-c226-11e8-8086-54e1ad486e52 
cifstestvol
```

#### Delete a volume
```
$ export X_CSI_SECRETS=admin_name="YOUR CIFS ADMIN USER",admin_password="YOUR CIFS ADMIN PASSWORD"

$ csc controller --endpoint tcp://127.0.0.1:10000 delete-volume csi-cifs-9bd0415d-c226-11e8-8086-54e1ad486e52
```

#### Get NodeID
```
$ csc node get-id --endpoint tcp://127.0.0.1:10000
CSINode
```
