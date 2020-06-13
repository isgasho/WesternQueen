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

