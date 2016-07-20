# Docker Machine driver for IIJ GIO P2PUB

Docker MachineのP2PUB向けドライバです。
自動でIIJ GIO P2PUBクラウドにDockerのホストを立てることができます。

[English](README.md)

## 動作環境

- Docker Machine 0.8+ (Docker Toolboxに同梱されています)

## インストール

### ソースからビルド

- go get -d -u github.com/iij/docker-machine-driver-p2pub
- go build -o /usr/local/bin/docker-machine-driver-p2pub github.com/iij/docker-machine-driver-p2pub/bin

### Windows

- [バイナリをダウンロード](https://github.com/iij/docker-machine-driver-p2pub/releases)
    - docker-machine-driver-p2pub-windows-amd64.exe
- 実行ファイルをdocker-machine.exeと同じフォルダにコピーして下さい

### Linux, OS X

- [バイナリをダウンロード](https://github.com/iij/docker-machine-driver-p2pub/releases)
    - docker-machine-driver-p2pub-linux-amd64 (Linux)
    - docker-machine-driver-p2pub-darwin-amd64 (OS X)
- install -c -m 755 docker-machine-driver-p2pub-linux-amd64 /usr/local/bin/docker-machine-driver-p2pub

## 使い方

```
[local]# export IIJAPI_ACCESS_KEY=<APIのアクセスキー>
[local]# export IIJAPI_SECRET_KEY=<APIのシークレットキー>
[local]# export GISSERVICECODE=<GISサービスコード>
[local]# docker-machine create -d p2pub p2machine
  :
```

オプション

| オプション | 環境変数 | 規定値 | 説明 |
|--------|--------|---------|-------------|
| `--p2pub-access-key` | `IIJAPI_ACCESS_KEY` | | APIアクセスキー(**必須**) |
| `--p2pub-secret-key` | `IIJAPI_SECRET_KEY` | | APIシークレットキー(**必須**) |
| `--p2pub-gis` | `GISSERVICECODE` | | P2(GIS)サービスコード(**必須**) |
| `--p2pub-server-type` | | VB0-1 | 仮想マシンのグレード -> [仮想サーバ品目](http://manual.iij.jp/p2/pubapi/59949011.html) |
| `--p2pub-server-group` | | | サーバグループ (`A` or `B`) |
| `--p2pub-system-storage` | | S30GB_CENTOS7_64 | システムストレージのグレード(OSを選択) -> [ストレージ品目](http://manual.iij.jp/p2/pubapi/59949023.html) |
| `--p2pub-data-storage` | | | 追加ストレージのグレード -> [ストレージ品目](http://manual.iij.jp/p2/pubapi/59949023.html) |
| `--p2pub-storage-group` | | | ストレージグループ (`Y` or `Z`) |
| `--p2pub-docker-port` | | 2376 | Dockerデーモンのポート番号 |
| `--p2pub-iba` | `IBASERVICECODE` | | システムストレージのサービスコード。指定なしなら契約を新規追加します |
| `--p2pub-ibb` | `IBBSERVICECODE` | | データストレージのサービスコード |
| `--p2pub-ivm` | `IVMSERVICECODE` | | 仮想マシンのサービスコード。指定なしなら契約を新規追加します |
| `--p2pub-private-only` | | | グローバルIPアドレスを付与せず、プライベートネットワークにDNSとゲートウェイ設定を追加する |

### Swarmクラスタの作り方

クラスタ作成

```
[local]# export IIJAPI_ACCESS_KEY=<APIのアクセスキー>
[local]# export IIJAPI_SECRET_KEY=<APIのシークレットキー>
[local]# export GISSERVICECODE=<GISサービスコード>
[local]# docker pull swarm
[local]# token=$(docker run --rm swarm create)
[local]# docker-machine create -d p2pub --swarm --master --swarm-discovery token://$token swarm-mng
  :
[local]# docker-machine create -d p2pub --swarm --swarm-discovery token://$tokne swarm-node01
  :
[local]# docker-machine create -d p2pub --swarm --swarm-discovery token://$tokne swarm-node02
  :
```

クラスタの利用

```
[local]# docker $(docker-machine config --swarm swarm-mng) version
[local]# docker $(docker-machine config --swarm swarm-mng) info
[local]# docker $(docker-machine config --swarm swarm-mng) ps
[local]# docker $(docker-machine config --swarm swarm-mng) pull alpine
[local]# docker $(docker-machine config --swarm swarm-mng) run alpine date
```

## Author

- Takashi WATANABE (@wtnb75)
