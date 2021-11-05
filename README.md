### README

#### clone 项目

`
git clone https://git-c.i.wxblockchain.com/PlatONE/src/node/platone-manager/platone-api-server.git
`
#### 安装mongodb

1. 安装前我们需要安装各个 Linux 平台依赖包

Ubuntu 18.04 LTS ("Bionic")/Debian 10 "Buster"：

`
sudo apt-get install libcurl4 openssl
`


Ubuntu 16.04 LTS ("Xenial")/Debian 9 "Stretch"：

`
sudo apt-get install libcurl3 openssl
`

2. 下载源码

MongoDB 源码下载地址：https://www.mongodb.com/download-center#community

`
wget https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu1604-4.2.8.tgz    # 下载
`

`
tar -zxvf mongodb-linux-x86_64-ubuntu1604-4.2.8.tgz                                    # 解压
`

`
mv mongodb-src-r4.2.8  /usr/local/mongodb4                          # 将解压包拷贝到指定目录
`

3. 环境配置

MongoDB 的可执行文件位于 bin 目录下，所以可以将其添加到 PATH 路径中：

`
export PATH=<mongodb-install-directory>/bin:$PATH
`

 <mongodb-install-directory> 为你 MongoDB 的安装路径。如本文的 /usr/local/mongodb4 
 
 `
 export PATH=/usr/local/mongodb4/bin:$PATH
 `
#### 启动api server
##### 1. 配置文件
根目录下config.toml文件包含了服务器配置，其描述如下：
```
[http]
ip = "0.0.0.0"   //api 服务器的ip
port = "9999"    //api 服务器的端口号
mode = "debug"  //模式
endpoint = "http://127.0.0.1:6791" //远程节点ip和端口号
restServer = "http://10.250.122.10:8000" //rest 服务器ip及端口号
capath = "/Users/night/dev/go/src/platone-api-server/model/monitor/keys/ca.cert" //ca文件路径，需手动指定本地ca路径


[db]
ip = "127.0.0.1" //mongo db的ip
port = "27017"  //mongo db的端口号
username = "root" //mongo db的用户名
password = "root" //mongo db的密码
dbname = "api-server" //mongo db name
```
##### 2. 启动流程

启动monitor server
```
cd platone-moniter/server
go run main.go
```

启动mongodb
```
sudo mongod
sudo mongo
```

创建账户：username：root password：root

构建数据库：构建data-manager

启动cmd/main.go

