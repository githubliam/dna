
<h1 align="center">DNA </h1>
<h4 align="center">Version 2.0 </h4>

[English](README.md) | 中文

欢迎来到DNA的源码库！ 

DNA致力于创建一个组件化、可自由配置、跨链支持、高性能、横向可扩展的区块链底层基础设施。 让部署及调用分布式应用变得更加非常简单。

常欢迎及希望能有更多的开发者加入到DNA中来。

## 特性

* 可扩展的轻量级通用智能合约
* 可扩展的WASM合约的支持
* 跨链交互协议（进行中）
* 多种加密算法支持 
* 高度优化的交易处理速度
* P2P连接链路加密(可选择模块)
* 多种共识算法支持 (VBFT/DBFT/SBFT/PoW/SOLO...)
* 快速的区块生成时间

## 目录

* [构建开发环境](#构建开发环境)
* [运行DNA](#运行DNA)
    * [测试模式](#测试模式)
* [使用示例](#使用示例)
	* [查询事务结果示例](#查询事务结果示例)
* [贡献代码](#贡献代码)
* [许可证](#许可证)

## 构建开发环境
成功编译DNA需要以下准备：

* Golang版本在1.9及以上
* 安装第三方包管理工具glide
* 正确的Go语言开发环境
* Golang所支持的操作系统


用make编译源码

```shell
$ make all
```

成功编译后会生成两个可以执行程序

* `DNA`: 节点程序/以命令行方式提供的节点控制程序
* `tools/sigsvr`: (可选)签名服务 - sigsvr是一个签名服务的server以满足一些特殊的需求。详细的文档可以在[这里](./docs/specifications/sigsvr_CN.md)参考

## 运行DNA

### 测试模式

在单机上创建一个目录，在目录下存放以下文件：
- 节点程序`DNA`
- 钱包文件`executor.dat` （注：`executor.dat`可通过`./DNA account add -d`生成）

使用命令 `$ ./DNA --testmode` 即可启动单机版的测试网络。

单机配置的例子如下：
- 目录结构

    ```shell
    $ tree
    └── node
        ├── DNA
        └── executor.dat
    ```

## 使用示例

### 查询事务示例

```shell
./DNA info status <TxHash>
```

如：

```shell
./DNA info status e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3
```

查询结果：
```shell
Transaction states:
{
   "TxHash": "e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3",
   "State": 1,
   "GasConsumed": 0,
   "Notify": [
      {
         "ContractAddress": "0200000000000000000000000000000000000000",
         "States": [
            "transfer",
            "ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",
            "AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce",
            95479777254
         ]
      }
   ]
}
```

## 贡献代码

请您以签过名的commit发送pull request请求，我们期待您的加入！
您也可以通过邮件的方式发送你的代码到开发者邮件列表，欢迎加入DNA邮件列表和开发者论坛。

另外，在您想为本项目贡献代码时请提供详细的提交信息，格式参考如下：

  Header line: explain the commit in one line (use the imperative)

  Body of commit message is a few lines of text, explaining things
  in more detail, possibly giving some background about the issue
  being fixed, etc etc.

  The body of the commit message can be several paragraphs, and
  please do proper word-wrap and keep columns shorter than about
  74 characters or so. That way "git log" will show things
  nicely even when it's indented.

  Make sure you explain your solution and why you're doing what you're
  doing, as opposed to describing what you're doing. Reviewers and your
  future self can read the patch, but might not understand why a
  particular solution was implemented.

  Reported-by: whoever-reported-it
  Signed-off-by: Your Name <youremail@yourhost.com>

## 许可证

DNA遵守Apache License, 版本2.0。 详细信息请查看项目根目录下的LICENSE文件。
