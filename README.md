## Basic Build Instructions

### Go

To build miner, you need a working installation of Go

### Build and install miner

1.Clone the repository:

```shell
git clone https://dataocean-backend.git
cd dataocean-backend/
```

2.Build miner

```shell
go build
#在dataocean文件夹下会生成 `dataoceanbackend` 可执行文件
```

### Start and run miner

```shell
./dataoceanbackend --minerAccount=cosmosxxxxxx --chainApi=xxxxx  --fileDir=xxxxx  --threshold=xxxxx  --aesKey=xxxxx

```
