# WesternQueen
阿里天池 首届云原生编程挑战赛 赛道一: 分布式统计和过滤的链路追踪


## 运行

* 主节点运行
```bash
./WesternQueen  -mode master
```
* 从节点运行

```bash
./WesternQueen  -mode slave1
./WesternQueen  -mode slave2
```

## 接口流水线

1. ready (common)
2. setParameter (common)
3. clientProcessData.Start(common -> slave)
4. readLines (slave -> 评测程序)
5. setWrongTraceId (slave -> master)
6. getWrongTrace (master -> slave)
7. sendCheckSum (master -> 评测程序)

## Docker测评命令

```bash
  docker login -u a2osdocker@1443039390876007 -p a2osdocker registry.cn-hangzhou.aliyuncs.com
  docker pull registry.cn-hangzhou.aliyuncs.com/a2os/tianchi:1.0
  docker run --rm -it  --net host -e "SERVER_PORT=8000" --name "clientprocess1" -d registry.cn-hangzhou.aliyuncs.com/a2os/tianchi:1.0
  docker run --rm -it  --net host -e "SERVER_PORT=8001" --name "clientprocess2" -d registry.cn-hangzhou.aliyuncs.com/a2os/tianchi:1.0
  docker run --rm -it  --net host -e "SERVER_PORT=8002" --name "backendprocess" -d registry.cn-hangzhou.aliyuncs.com/a2os/tianchi:1.0


docker pull registry.cn-hangzhou.aliyuncs.com/cloud_native_match/scoring:0.1
docker run --rm --net host -e "SERVER_PORT=8081" --name scoring -d registry.cn-hangzhou.aliyuncs.com/cloud_native_match/scoring:0.1
```