# プライベートネットワーク内でswarmクラスタを動かす

docker-machine-driver-p2pubでプライベートネットワークにdocker-swarmクラスタを構築する例を紹介します。

## 踏み台サーバを1個作る

ローカルのdocker-machineでP2PUBに踏み台サーバを1台作ります。

```
[local]# docker-machine create -d p2pub base
  :
```

## 踏み台サーバにdocker-machine一式をインストール

```
[local]# docker-machine ssh base
[root@base]# wget https://github.com/iij/docker-machine-driver-p2pub/releases/download/0.9/docker-machine-driver-p2pub-linux-amd64
[root@base]# install -c -m 755 docker-machine-driver-p2pub-linux-amd64 /usr/bin/docker-machine-driver-p2pub
[root@base]# wget https://github.com/docker/machine/releases/download/v0.8.0/docker-machine-Linux-x86_64
[root@base]# install -c -m 755 docker-machine-Linux-x86_64 /usr/bin/docker-machine
```

## 踏み台サーバを設定する

docker-machineの処理の一部でyum updateをしたり、docker.comからdockerのバイナリをインストールしたりしますので、完全にプライベートネットワークのみではインストールが進みません。
踏み台サーバはインターネットとプライベートネットワークの両方につながり、プライベートネットワークに接続されたswarmクラスタを構成するマシンからNATで外につながるように設定します。

```
[root@base]# firewall-cmd --zone=internal --change-interface=eth1
[root@base]# firewall-cmd --zone=public --add-masquerade
[root@base]# firewall-cmd --zone=internal --change-interface=eth1 --permanent
[root@base]# firewall-cmd --zone=public --add-masquerade --permanent
```

ここで踏み台のDNS設定と、プライベート側のIPアドレスをメモしておきます。あとで使います。

```
[root@base]# PRIVATEIP=$(ip addr show dev eth1 | awk '/inet/{print $2;exit;}' | cut -f1 -d/)
[root@base]# DNS=$(grep nameserver /etc/resolv.conf  | head -n 1 | cut -f2 -d' ')
```

# docker-machineを使ったprivate swarmのデプロイ

踏み台サーバで作業します。

## swarmのtokenを取得

```
[root@base]# docker pull swarm
[root@base]# token=$(docker run --rm swarm create)
```

## swarm-managerのインストール

`--p2pub-private-only`オプションの引数には、ゲートウェイアドレスとDNSのアドレスをカンマ(`,`)で区切って与えます。

```
[root@base]# docker-machine create -d p2pub --p2pub-private-only $PRIVATEIP,$DNS --swarm --swarm-master --swarm-discovery token://${token} swarm-mng
Running pre-create checks...
Creating machine...
(swarm-mng) creating VM
  :
Docker is up and running!
To see how to connect your Docker Client to the Docker Engine running on this virtual machine, run: docker-machine env swarm-mng
```

## swarm-nodeのインストール

以下の例では、2台のノードにインストールしています。

```
[root@base]# for i in swarm-node00 swarm-node01 ; do docker-machine create -d p2pub --p2pub-private-only $PRIVATEIP,$DNS --swarm --swarm-discovery token://${token} $i ; done
  :
```

# private swarmを使ってみよう

踏み台のdockerから使います。

```
[root@base]# docker $(docker-machine config --swarm swarm-mng) ps
[root@base]# docker $(docker-machine config --swarm swarm-mng) info
[root@base]# docker $(docker-machine config --swarm swarm-mng) version
```
