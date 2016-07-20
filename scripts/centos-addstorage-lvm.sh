#! /bin/sh

adddev=${1-/dev/vdb}

[ -n "$adddev" ] || exit 1
[ -b $adddev ] || exit 1

# stopping docker
systemctl stop docker

# create thin pool
pvcreate $adddev
vgcreate docker $adddev
lvcreate --wipesignatures y -n thinpool docker -l 95%VG
lvcreate --wipesignatures y -n thinpoolmeta docker -l 1%VG
lvconvert -y --zero n -c 512K --thinpool docker/thinpool --poolmetadata docker/thinpoolmeta
cat <<EOF > /etc/lvm/profile/docker-thinpool.profile
activation {
  thin_pool_autoextend_threshold=80
  thin_pool_autoextend_percent=20
}
EOF
lvchange --metadataprofile docker-thinpool docker/thinpool
lvs -o+seg_monitor

# remove old
rm -rf /var/lib/docker/

# setup service
cat <<EOF > /etc/docker/daemon.json
{
  "storage-opts": [
    "dm.thinpooldev=/dev/mapper/docker-thinpool",
    "dm.use_deferred_removal=true"
  ]
}
EOF
systemctl daemon-reload

# starting docker
systemctl start docker
