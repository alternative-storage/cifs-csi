FROM centos:7

LABEL maintainers="Kenjiro Nakayama"
LABEL description="CIFS CSI Plugin"

RUN yum -y update && yum install -y samba-client && \
    yum -y clean all

COPY cifsplugin /cifsplugin

RUN chmod +x /cifsplugin && \
    mkdir -p /var/log/cifs

ENTRYPOINT ["/cifsplugin"]
