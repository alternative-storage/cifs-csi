FROM ubuntu:16.04
LABEL maintainers="Kenjiro Nakayama"
LABEL description="CIFS CSI Plugin"

ENV CEPH_VERSION "mimic"

RUN apt-get update && apt-get install -y wget && \
    wget -q -O- 'https://download.ceph.com/keys/release.asc' | apt-key add - && \
    echo "deb http://download.ceph.com/debian-$CEPH_VERSION/ xenial main" | tee /etc/apt/sources.list.d/ceph-$CEPH_VERSION.list && \
    apt-get update && apt-get install -y samba-client --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

COPY cifsplugin /cifsplugin

RUN chmod +x /cifsplugin && \
    mkdir -p /var/log/cifs

ENTRYPOINT ["/cifsplugin"]
