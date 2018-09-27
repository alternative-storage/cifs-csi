
Configure in smb.conf
---

~~~
[global]
        add share command = /usr/local/bin/addshare.sh
        delete share command = /usr/local/bin/delshare.sh
~~~

Deploy scripts
---

~~~
sudo curl https://raw.githubusercontent.com/alternative-storage/cifs-csi/master/examples/samba/addshare.sh -o /usr/local/bin/addshare.sh
sudo curl https://raw.githubusercontent.com/alternative-storage/cifs-csi/master/examples/samba/delshare.sh -o /usr/local/bin/delshare.sh
~~~

Preparation
---

~~~
// You might want to add `admin users = admin` in smb.conf rather than root user.
smbpasswd -a root

// Default samba's user share
mkdir -p mkdir -p /var/lib/samba/usershares
~~~


~~~
[global]
        usershare max shares = 100
~~~
