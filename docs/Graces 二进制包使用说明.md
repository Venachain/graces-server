# README

## 环境准备

### 硬件环境准备

- 主  机：Intel Xeon E5-2650或以上。

- 内  存：16G或以上。

- 硬  盘：32GB或以上。

- 图形卡：VGA/DVI。

### 软件环境准备

- OS	Linux 64bit
- Golang 1.14.4 或以上
- Venachain v1.0.0 或以上
- MongoDB 4.2.8 或以上
- npm 6.14.13 或以上
- nodjs v14.17.1 或以上
- Graces v1.0.0

## 安装步骤

1. 下载、安装和配置 Golang

2. 下载、安装和启动 MongoDB

3. 下载 Graces V1.0.0

    - 下载 Graces 二进制包：

      - linux 下载地址：https://github.com/Venachain/graces-server/releases/download/v1.0.0/venachain_graces_linux_amd64.tar.gz
      - macos 下载地址：https://github.com/Venachain/graces-server/releases/download/v1.0.0/venachain_graces_macos_arm64.tar.gz

      下载完成后对项目文件进行解压：

       ```sh
       tar -zxvf venachain_graces_linux_amd64.tar.gz
       ```
      
    - 解压完成后会得到如下几个文件：
   
      - graces-server：这是 graces-server 的二进制包，可以直接运行
      
      - config.toml：这是 graces-server 的配置文件，如有需要可以修改里面的配置信息
      
      - dist：这是 graces-web 的二进制包，将其放到 Nginx 上运行即可。

4. 配置 MongoDB

   MongoDB 启动成功后，需要为 graces 创建数据库和用户。

    1. 先用自己设置好的账号密码登录 MongoDB。

    2. 使用下面命令创建 graces 数据库，并进入 graces 数据库。

   ```shell
    use graces;
   ```

    3. 使用下面命令在 graces 数据库下创建 test 用户并赋予读写和管理角色。

   ```shell
    db.createUser(
       {
        user: "test",
        pwd: "test",
        roles: [ "readWrite", "dbAdmin" ]
      }
    );
   ```   
   完成上述步骤，我们就可以在 graces-server 中使用 test 用户去操作 MongoDB 的 graces 数据库了。

   注：test 只是一个示例用户，您可以换成自己想要创建的用户和密码。


6. 配置 graces-web

   进到 dist/static 目录下，将 `config.json` 文件中的 `localhost:9999` 修改为 `graces-server` 所在机器的 IP 端口号或域名端口号，如果不修改则默认使用 `localhost:9999` 。

   ``` json
    {
     "BASE_URL": "http://localhost:9999/api",
     "BASE_WS": "ws://localhost:9999/api",
     "BASE_ENV": "production"
    }
   ```

7. 配置 graces-server

   1. 在 `config.toml` 文件中配置 graces-server 运行的 IP 地址和端口号。需要注意的是，cors 的值必须是 graces-web 的运行地址，本示例 graces-web 运行在 http://localhost:8080 中。

       ```toml
       [http]
       ip = "127.0.0.1"
       port = "9999"
       # mode 必须是 "release"、"debug"、"test" 中的一个
       mode = "debug"
       # cors 跨域资源共享白名单
       cors = "http://localhost:8080"
       ```

   2. 在 config.toml 文件中配置 graces-server 所需的 MongoDB 信息，其中 username 和 password 应该填写为上面我们已经在 MongoDB 中配置好的 graces 数据库的用户和密码。

      ```toml
      [db]
      ip = "127.0.0.1"
      port = "27017"
      username = "test"
      password = "test"
      dbname = "graces"
      timeout = 10
      ```

8. 启动 Graces

    1. 启动 graces-server

       进到 graces-server 所在的目录下，执行以下命令

       ```sh
       nohup ./graces-server > ./graces.log 2>&1 &
       ```

    2. 启动 graces-web

       将 dist 文件夹放到 nginx 服务器中，通过 nginx 来启动 graces-web

9. 访问 Graces

   开打浏览器，输入 `http://localhost:8080` 便可以进到 Graces 主页面。

   ![](docs/imgs/index.png)

## 问题处理

1. 关于跨域问题的处理

   注意：如果 graces-server 和 graces-web 不是部署在同一机器上，则 graces-web 与 graces-server 的链接可能会出现跨域问题，这时候需要做两步操作:

    1. 进到 dist/static 目录下，将 `config.json` 文件中的 `localhost:9999` 修改为 `graces-server` 所在机器的 IP 端口号或域名端口号，整体内容如下：

   ``` json
    {
     "BASE_URL": "http://localhost:9999/api",
     "BASE_WS": "ws://localhost:9999/api",
     "BASE_ENV": "production"
    }
   ```

    2. 在 `graces-server` 里面找到 `config.toml` 配置文件，修改 ip 的值为 `graces-server` 所在机器的公网 ip，修改 cors 的值为 `graces-web` 的访问地址，如下：

       ```toml
       [http]
       ip = "graces-server 的公网ip"
       port = "9999"
       # mode 必须是 "release"、"debug"、"test" 中的一个
       mode = "debug"
       # cors 跨域资源共享白名单，这是 graces-web 的访问地址，如：http://localhost:8080
       cors = "graces-web 的访问地址"
       ```
    
2. mongodb 链接失败

   如果在启动 graces-server 过程中出现以下错误，则需要确认一下 graces-server 连接 mongodb 的账号密码配置是否正确。

   ```shell
   FATA[0000] failed to connection DB： connection() error occured during connection handshake: auth error: sasl conversation error: unable to authenticate using mechanism "SCRAM-SHA-1": (AuthenticationFailed) Authentication failed.
    ```