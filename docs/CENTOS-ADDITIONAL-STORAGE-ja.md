# [CentOS7]追加ストレージの使い方

docker-machineの`--p2pub-data-storage <ストレージのグレード>`や`--p2pub-ibb <ストレージのサービスコード>`では、ストレージがVMにアタッチされるだけでdockerが使うところまでは設定されません。

docker-machineでデプロイした直後の状態では、CentOSではloopbackデバイスをdm-thinで使いますが、このモードでは性能上の問題があるため、テスト用途以外では推奨されていません。docker infoでも警告が表示されます。

```
[root@loopback ~]# docker info
  :
Storage Driver: devicemapper
 Pool Name: docker-253:1-33676068-pool
   :
 Data loop file: /var/lib/docker/devicemapper/devicemapper/data
 WARNING: Usage of loopback devices is strongly discouraged for production use. Either use `--storage-opt dm.thinpooldev` or use `--storage-opt dm.no_warn_on_loop_devices=true` to suppress this warning.
 Metadata loop file: /var/lib/docker/devicemapper/devicemapper/metadata
 Library Version: 1.02.107-RHEL7 (2016-06-09)
   :
```

CentOSでこれを避けるには、以下の3つのやり方があります。

- direct-lvm
    - loopbackを使わず、本物のデバイスを使ってthinpoolを構築する
- btrfs
    - 安定性には欠ける
    - btrfsのsubvolumeとsnapshotの機能を使う
- overlayfs
    - ファイルシステムを差分で管理できるunion filesystemの一種であるoverlayfsを使う

## direct-lvm

Dockerのドキュメントでは[Docker and the Device Mapper storage driver](https://docs.docker.com/engine/userguide/storagedriver/device-mapper-driver/)で説明されています。

### dm-thinデバイスの設定

`--p2pub-data-storage`や`--p2pub-ibb`で設定した追加ストレージは/dev/vdbにアタッチされた状態になります。/dev/vdbでthin poolを設定します。

```
[root@machine]# pvcreate /dev/vdb
  Physical volume "/dev/vdb" successfully created
[root@machine]# vgcreate docker /dev/vdb
  Volume group "docker" successfully created
[root@machine]# lvcreate --wipesignatures y -n thinpool docker -l 95%VG
  Logical volume "thinpool" created.
[root@machine]# lvcreate --wipesignatures y -n thinpoolmeta docker -l 1%VG
  Logical volume "thinpoolmeta" created.
[root@machine]# # lvconvert -y --zero n -c 512K --thinpool docker/thinpool --poolmetadata docker/thinpoolmeta
  WARNING: Converting logical volume docker/thinpool and docker/thinpoolmeta to pool's data and metadata volumes.
  THIS WILL DESTROY CONTENT OF LOGICAL VOLUME (filesystem etc.)
  Converted docker/thinpool to thin pool.
```

vi /etc/lvm/profile/docker-thinpool.profile

以下の内容を入力
```
activation {
  thin_pool_autoextend_threshold=80
  thin_pool_autoextend_percent=20
}
```

```
[root@machine]# lvchange --metadataprofile docker-thinpool docker/thinpool
  Logical volume "thinpool" changed.
[root@machine]# lvs -o+seg_monitor
  LV       VG     Attr       LSize  Pool Origin Data%  Meta%  Move Log Cpy%Sync Convert Monitor
  thinpool docker twi-a-t--- 95.00g             0.00   0.01                             monitored
```

ここまででdocker用のthinpoolができました。

### dockerの設定

docker daemonを停止し、/var/lib/dockerを削除します。

```
[root@machine]# systemctl stop docker
[root@machine]# rm -rf /var/lib/docker
```

dockerの追加オプションを設定します。

vi /etc/docker/daemon.json

以下の内容を入力
```js
{
  "storage-opts": [
    "dm.thinpooldev=/dev/mapper/docker-thinpool",
    "dm.use_deferred_removal=true"
  ]
}
```

### dockerの起動と動作確認

```
[root@machine]# systemctl start docker
[root@machine]# docker info
  :
Storage Driver: devicemapper
 Pool Name: docker-thinpool  ← プール名を確認
 Pool Blocksize: 524.3 kB
 Base Device Size: 10.74 GB
 Backing Filesystem: xfs
 Data file:
 Metadata file:
 Data Space Used: 20.45 MB
 Data Space Total: 102 GB ← データデバイスの容量と比較
 Data Space Available: 102 GB
 Metadata Space Used: 147.5 kB
 Metadata Space Total: 1.07 GB
 Metadata Space Available: 1.069 GB
 Udev Sync Supported: true
 Deferred Removal Enabled: true
 Deferred Deletion Enabled: false
 Deferred Deleted Device Count: 0
 Library Version: 1.02.107-RHEL7 (2016-06-09)
 　:
```

### スクリプト

参考までに、設定スクリプトを[scripts/centos-addstorage-lvm.sh](../scripts/centos-addstorage-lvm.sh)に置かせていただきました。

## Btrfs

Btrfsを使うこともできます。

Dockerのドキュメントでは[Docker and Btrfs in practice](https://docs.docker.com/engine/userguide/storagedriver/btrfs-driver/)で説明されています。

### docker daemonの停止

```
[root@machine]# systemctl stop docker
```

### btrfsのファイルシステム作成とマウント

ファイルシステムとマウントポイントの作成
```
[root@machine]# mkfs.btrfs /dev/vdb
btrfs-progs v3.19.1
See http://btrfs.wiki.kernel.org for more information.

Turning ON incompat feature 'extref': increased hardlink limit per file to 65536
Turning ON incompat feature 'skinny-metadata': reduced-size metadata extent refs
fs created label (null) on /dev/vdb
	nodesize 16384 leafsize 16384 sectorsize 4096 size 100.00GiB
[root@machine]# rm -rf /var/lib/docker
[root@machine]# mkdir /var/lib/docker
```

vi /etc/fstab

以下を追記
```
/dev/vdb /var/lib/docker btrfs defaults 0 0
```

マウント
```
[root@machine]# mount /dev/vdb /var/lib/docker
```

vi /etc/systemd/system/docker.service

ExecStart=の行の`--storage-driver devicemapper`を`--storage-driver btrfs`に書き換え

```
[root@machine]# systemctl daemon-reload
```

### docker daemonの起動と動作確認

```
[root@machine]# systemctl docker start
[root@machine]# docker info
  :
Storage Driver: btrfs
 Build Version: Btrfs v3.19.1
 Library Version: 101
  :
```

## overlayfs

3つの中で最も手軽なものがoverlayfsを使う方法でしょう。

### docker daemonの停止

```
[root@machine]# systemctl stop docker
```

### ファイルシステムの作成とマウント

```
[root@machine]# mkfs.ext4 /dev/vdb
mke2fs 1.42.9 (28-Dec-2013)
Filesystem label=
[Unit]
OS type: Linux
Block size=4096 (log=2)
Fragment size=4096 (log=2)
Stride=0 blocks, Stripe width=0 blocks
6553600 inodes, 26214400 blocks
1310720 blocks (5.00%) reserved for the super user
First data block=0
Maximum filesystem blocks=2174746624
800 block groups
32768 blocks per group, 32768 fragments per group
8192 inodes per group
Superblock backups stored on blocks:
	32768, 98304, 163840, 229376, 294912, 819200, 884736, 1605632, 2654208,
	4096000, 7962624, 11239424, 20480000, 23887872

Allocating group tables: done
Writing inode tables: done
Creating journal (32768 blocks): done
Writing superblocks and filesystem accounting information: done
```

vi /etc/fstab

以下を追記
```
/dev/vdb /var/lib/docker ext4 defaults 0 0
```

マウント
```
[root@machine]# mount /dev/vdb /var/lib/docker
```

vi /etc/systemd/system/docker.service

ExecStart=の行の`--storage-driver devicemapper`を`--storage-driver overlay`に書き換え

```
[root@machine]# systemctl daemon-reload
```

### docker daemonの起動と動作確認

```
[root@machine]# systemctl start docker
[root@machine]# docker info
  :
Storage Driver: overlay
 Backing Filesystem: extfs
  :
```

### 追加ディスクの有無とは無関係

追加ディスクを使わなくてもoverlayfsを使うことはできます。その場合はstorage-driverの書き換えだけで済みます。

docker-machineコマンドでも、以下のように`--engine-storage-driver overlay`とオプションを追加すると、最初からoverlayfsを使うようになります。

```
[local]# docker-machine create -d p2pub --engine-storage-driver overlay
```

### overlayfsの注意点

overlayfsは一部POSIX非互換があるため、rpmやyum等が正常に動作しないケースがあるようです。また、overlayfsにはinodeの消費量が多くなるという問題があります。その点を改善したoverlayfs2という次期版もありますが、CentOS7ではまだ使えません。
