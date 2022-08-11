增加了consensus/bihs
在params/config.go中增加了BIHSconfig,params的chainConfig也增加了一个字段，同时若干参数数量不匹配也需要错误修复
eth/ethconfig/config.go 中CreateConsensusEngine函数
eth/backend.go中的代码也要改，增加了一段bihs初始化代码
miner的worker.go代码case <-timer.C:也要改
miner和worker都添加了一些函数
关于gensis的配置文件


# 关于拜占庭场景通过投票达成共识的本质思考

为简化讨论，假设共有3f+1个节点，其中f个恶意节点，2f+1个诚实节点，并且这里只讨论对单个值的共识（多次迭代，即可实现对多个值的共识）。

这里假设投票模型是按view切换，每个view内有若干轮，每轮内诚实节点最多只会投1票。（篇幅原因不赘述了，类似pbft）

然后有个很关键的结论（引理1）：每轮最多只有1个值能获得2f+1个投票。

证明略。

首先定义下锁定：每个view有一特定轮投票，在该轮对某个值投票，即锁定在该值。

根据引理1，该轮最多只有一个值能被2f+1个节点锁定。因此如果有2f+1个节点在某个view锁定了某个值，便可认为在这个view达成了共识；如果有某种机制保证到了下个view，这个值仍将继续被锁定，那么可以认为彻底达成了共识。

于是对锁定后的投票加个限制，一旦某个节点锁定在某个值，那么它只对这个值投票（此即安全性），除非得知2f+1个其他节点投了更新（view更高）的某个值（此即活性，解锁）。这样便能保证一旦2f+1个节点锁定在了某个值，即可认为彻底达成了共识。简而言之：共识=2f+1锁定，这是满足安全性的，接下来是活性。

又有个很关键的结论（引理2）：如果锁定蕴含对之前view锁定值的解锁能力，那么在网络良好时，只要锁定最新值的节点作为leader，并且所有诚实节点都在同一个view，那么便可达成共识。

证明略。

所以，只要让锁定蕴含解锁能力，活性便也解决了。

于是只需要复用引理1，在锁定轮之前先进行一轮投票，只有得到2f+1个投票的值才是合法的锁定值。


未完待续




## 启动说明

### 1. 编译二进制文件
```buildoutcfg
make all
```

在`./build/bin`目录下可以看到下列二进制文件
```buildoutcfg
abidump
abigen
bootnode
checkpoint-admin
clef
devp2p
ethkey
evm
faucet
geth
p2psim
puppeth
rlpdump

```

### 2.环境准备
以本地启动四个节点为例
```
node1
node2
```
将geth 拷贝到每个文件夹下，执行：
```buildoutcfg
./geth --datadir ./data/ init bihs_genesis.json 
```

该命令初始化`genesis`为`bihs`共识，并生成`p2p nodekey`在`data/geth`目录下，需要手动计算私钥的地址，并填入到https://github.com/zhiqiangxu/go-ethereum/blob/web3q_bihs/consensus/bihs/gov/gov.go#L20 ，该模块内部以round robin方式确定每一轮的leader，仅用于demo。

### 3. 启动节点
假定二个节点的p2p端口为2000 ~ 2001,我们暂时只开放node1的http端口，默认为8545
首先通过下面的命令打印各个节点的公钥
```buildoutcfg
bootnode -nodekey ./data/geth/nodekey -writeaddress
```
将节点列表放入<datadir>/geth/static-nodes.json文件中，让节点主动连接和重连：https://geth.ethereum.org/docs/interface/peer-to-peer

```buildoutcfg
[
    "enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000",
    "enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001"
]
```

启动命令
```
./geth  --datadir ./data --networkid 121 --port 2000 --http --http.addr 0.0.0.0 --http.port 8545  --authrpc.port=8551 --miner.gasprice 0 --mine --miner.etherbase=0xde5B5Dd07C7EE63712b334EcD59E3FA173E6d56E --syncmode full --nodiscover --verbosity 5  --authrpc.port=8551
./geth  --datadir ./data --networkid 121 --port 2001 --authrpc.port=8552 --miner.gasprice 0 --mine --miner.etherbase=0xD642f9b4c28F6bA62126144B7E26e8Cf85CB2d3a --syncmode full --nodiscover --verbosity 5
```
其中`miner.etherbase`需要跟p2p私钥`geth/nodekey`对应。