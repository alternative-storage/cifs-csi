[![Build Status](https://travis-ci.org/alternative-storage/cifs-csi.svg?branch=master)](https://travis-ci.org/alternative-storage/cifs-csi)

# CSI driver for CIFS

[Container Storage Interface (CSI)](https://github.com/container-storage-interface/) driver, provisioner, and attacher for CIFS (SMB, Samba, Windows Share) network filesystems.



## Test
Get ```csc``` tool from https://github.com/rexray/gocsi/tree/master/csc

#### Get plugin info
```
$ csc identity plugin-info --endpoint tcp://127.0.0.1:10000
"csi-cifsplugin"	"0.3.0"
```

#### NodeStage a volume
```
$ export CIFS_SERVER="Your Server IP (Ex: 10.10.10.10)"
$ export CIFS_SHARE="Your NFS share"
$ export X_CSI_SECRETS=userID=test,userKey=password

$ csc node stage --endpoint tcp://127.0.0.1:10000  --attrib server=$CIFS_SERVER --attrib share=$CIFS_SHARE cifstestvol
cifstestvol
```

#### NodePublish a volume

**NOTE**: You must stage a volume by above step.

```
$ export CIFS_SERVER="Your Server IP (Ex: 10.10.10.10)"
$ export CIFS_SHARE="Your NFS share"
$ csc node publish --endpoint tcp://127.0.0.1:10000 --target-path /mnt/cifs --attrib server=$CIFS_SERVER --attrib share=$CIFS_SHARE cifstestvol
cifstestvol
```

#### NodeUnpublish a volume
```
$ csc node unpublish --endpoint tcp://127.0.0.1:10000 --target-path /mnt/cifs cifstestvol
cifstestvol
```

#### Get NodeID
```
$ csc node get-id --endpoint tcp://127.0.0.1:10000
CSINode
```
