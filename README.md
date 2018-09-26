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
Get ```csc``` tool from https://github.com/rexray/gocsi/tree/master/csc

#### Get plugin info
```
$ csc identity plugin-info --endpoint tcp://127.0.0.1:10000
"csi-cifsplugin"	"0.3.0"
```

#### Create a volume
```
$ export X_CSI_SECRETS=admin_name="YOUR CIFS ADMIN USER",admin_password="YOUR CIFS ADMIN PASSWORD"

$ csc controller --endpoint tcp://127.0.0.1:10000 create-volume testvol --params server=$CIFS_SERVER --params path="/tmp"
```


#### NodeStage a volume
```
$ export CIFS_SERVER="Your Server IP (Ex: 10.10.10.10)"
$ export CIFS_SHARE="Your CIFS share"
$ export X_CSI_SECRETS=username="YOUR CIFS MOUNT USER",password="YOUR PASSWORD"

$ csc node stage --endpoint tcp://127.0.0.1:10000  --attrib server=$CIFS_SERVER --attrib share=$CIFS_SHARE --staging-target-path=/mnt/cifs cifstestvol
cifstestvol
```

#### NodePublish a volume

**NOTE**: You must stage a volume by above step beforehand.

```
$ export CIFS_SERVER="Your Server IP (Ex: 10.10.10.10)"
$ export CIFS_SHARE="Your CIFS share"
$ csc node publish --endpoint tcp://127.0.0.1:10000 --staging-target-path /mnt/cifs --target-path /mnt/cifs-bind cifstestvol
cifstestvol
```

#### NodeUnpublish a volume
```
$ csc node unpublish --endpoint tcp://127.0.0.1:10000 --target-path /mnt/cifs-bind cifstestvol
cifstestvol
```

#### NodeUnstage a volume
```
$ csc node unstage --endpoint tcp://127.0.0.1:10000 --staging-target-path /mnt/cifs cifstestvol
cifstestvol
```

#### Delete a volume
```
$ export X_CSI_SECRETS=admin_name="YOUR CIFS ADMIN USER",admin_password="YOUR CIFS ADMIN PASSWORD"

$ csc controller --endpoint tcp://127.0.0.1:10000 delete-volume
```


#### Get NodeID
```
$ csc node get-id --endpoint tcp://127.0.0.1:10000
CSINode
```
