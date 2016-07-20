# Docker Machine driver for IIJ GIO P2PUB

This is the P2PUB driver plugin for docker-machine.
It allows you to automate provisioning of Docker hosts on IIJ GIO P2PUB cloud.

[日本語](README-ja.md)

## Requirements

- Docker Machine 0.8+ (included in Docker Toolbox)

## Installation

### Build from Source

- go get -d -u github.com/iij/docker-machine-driver-p2pub
- go build -o /usr/local/bin/docker-machine-driver-p2pub github.com/iij/docker-machine-driver-p2pub/bin

### Windows

- [Download release](https://github.com/iij/docker-machine-driver-p2pub/releases)
    - docker-machine-driver-p2pub-windows-amd64.exe
- copy executable to the same folder as docker-machine.exe

### Linux, OS X

- [Download release](https://github.com/iij/docker-machine-driver-p2pub/releases)
    - docker-machine-driver-p2pub-linux-amd64 (Linux)
    - docker-machine-driver-p2pub-darwin-amd64 (OS X)
- install -c -m 755 docker-machine-driver-p2pub-linux-amd64 /usr/local/bin/docker-machine-driver-p2pub

## Usage

```
[local]# export IIJAPI_ACCESS_KEY=<your access key>
[local]# export IIJAPI_SECRET_KEY=<your secret key>
[local]# export GISSERVICECODE=<your gis service code>
[local]# docker-machine create -d p2pub p2machine
  :
```

Options

| Option | EnvVar | Default | description |
|--------|--------|---------|-------------|
| `--p2pub-access-key` | `IIJAPI_ACCESS_KEY` | | API Access Key(**required**) |
| `--p2pub-secret-key` | `IIJAPI_SECRET_KEY` | | API Secret Key(**required**) |
| `--p2pub-gis` | `GISSERVICECODE` | | P2(GIS) Service Code(**required**) |
| `--p2pub-server-type` | | VB0-1 | Grade of Virtual Machine -> [仮想サーバ品目](http://manual.iij.jp/p2/pubapi/59949011.html) |
| `--p2pub-server-group` | | | Server Group (`A` or `B`) |
| `--p2pub-system-storage` | | S30GB_CENTOS7_64 | Type of System Storage(Operating System) -> [ストレージ品目](http://manual.iij.jp/p2/pubapi/59949023.html) |
| `--p2pub-data-storage` | | | Grade of Additional Storage -> [ストレージ品目](http://manual.iij.jp/p2/pubapi/59949023.html) |
| `--p2pub-storage-group` | | | Storage Group (`Y` or `Z`) |
| `--p2pub-docker-port` | | 2376 | Port Number of Docker daemon |
| `--p2pub-iba` | `IBASERVICECODE` | | System Storage Service Code. add new if not specified |
| `--p2pub-ibb` | `IBBSERVICECODE` | | Data Storage Service Code |
| `--p2pub-ivm` | `IVMSERVICECODE` | | Virtual Machine Service Code. add new if not specified |
| `--p2pub-private-only` | | | don't attach global IP and set up DNS/Gateway for private network |

### create Swarm cluster

Build Cluster

```
[local]# export IIJAPI_ACCESS_KEY=<your access key>
[local]# export IIJAPI_SECRET_KEY=<your secret key>
[local]# export GISSERVICECODE=<your gis service code>
[local]# docker pull swarm
[local]# token=$(docker run --rm swarm create)
[local]# docker-machine create -d p2pub --swarm --master --swarm-discovery token://$token swarm-mng
  :
[local]# docker-machine create -d p2pub --swarm --swarm-discovery token://$tokne swarm-node01
  :
[local]# docker-machine create -d p2pub --swarm --swarm-discovery token://$tokne swarm-node02
  :
```

Use Cluster

```
[local]# docker $(docker-machine config --swarm swarm-mng) version
[local]# docker $(docker-machine config --swarm swarm-mng) info
[local]# docker $(docker-machine config --swarm swarm-mng) ps
[local]# docker $(docker-machine config --swarm swarm-mng) pull alpine
[local]# docker $(docker-machine config --swarm swarm-mng) run alpine date
```

## Author

- Takashi WATANABE (@wtnb75)
