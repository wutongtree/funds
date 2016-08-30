## 设计实现

funds基于Hypperledger实现的基金管理，Hyperledger为我们提供了如下的功能：

* 用户管理

Hyperledger的membersrvc模块提供了基本的用户管理功能，基于PKI体系的用户系统保证了交易的安全性。用户管理本身采用配置文件进行初始化，我们会进行一些扩展。

* 共识算法

共识算法提供了在分布式环境下解决数据一致性问题的方法。

* 区块链存储

区块链存储把所有的交易结果都存储在区块链上，称为ledger，任何人都可以查询ledger上的信息。

### 架构设计

架构设计包含三大部分：web client、App、Hyperledger。如下图

fund架构图.jpg

web client：提供对外操作UI，实现user的输入输出简单处理后向App发送http request并接收response。
App：连接client与Hyperledger的中间层，负责接收client的httprequest，将request数据整理打包后通过Hyperledger提供API发送给Hyperledger处理；Hyperledger处理完成后返回处理结果给App，并有App包装后返回给client。
Hyperledger：基金管理系统底层区块链技术实现，提供memberSrv服务、peer共识服务、chaincode服务。负责执行交易并将交易相关信息保存于Ledger中。

###数据结构及数据流


