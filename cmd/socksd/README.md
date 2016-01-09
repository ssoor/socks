# SOCKS
[![Build Status](https://travis-ci.org/eahydra/socks.svg?branch=master)](https://travis-ci.org/eahydra/socks)

cmd/socksd 编译时需要 SOCKS 项目支持 , 目前这个项目支持的加密方式有: rc4, des, aes-128-cfb, aes-192-cfb and aes-256-cfb, 上游服务器可以使用: shadowsocks, socsk5.

# 安装
如果你有一个 go 语言开发环境, 你可以通过执行以下命令通过源代码来安装.
```
go get github.com/eahydra/socks/cmd/socksd
```

# 使用
配置文件使用 Json 格式. 这个文件必须命名为 **socks.config** 并且和生成的执行文件放在一起.
配置文件内容:
```json
{
  "pac": {
    "address": "127.0.0.1:2016",
    "upstream": {
      "type": "shadowsocks",
      "crypto": "aes-128-cfb",
      "password": "111222333",
      "address": "127.0.0.1:1080"
    },
    "rules": [
      {
        "name": "local_proxy",
        "proxy": "127.0.0.1:2333",
        "socks4": "127.0.0.1:2334",
        "socks5": "127.0.0.1:2335",
        "local_rule_file": "Hijacker.txt",
        "remote_rule_file": "httpss://raw.githubusercontent.com/Leask/BRICKS/master/gfw.bricks"
      }
    ]
  },
  "proxies": [
    {
      "http": ":2333",
      "socks4": ":2334",
      "socks5": ":2335",
      "upstreams": [
        {
          "type": "shadowsocks",
          "crypto": "aes-128-cfb",
          "password": "111222333",
          "address": "127.0.0.1:1080"
        }
      ]
    }
  ]
}

```

*  **pac**	- PAC 配置信息
    * **address**   - PAC服务器监听信息 (127.0.0.1:50000)
    * **upstream**  - (OPTIONAL)  读取 **remote_rule_file** 使用的代理信息
    * **rules**     - **rules** 数组, PAC服务器运行使用的规则信息

* **rules** - PAC解析规则信息
    * **name**     - (OPTIONAL)  PAC 名称
    * **proxy**    - (OPTIONAL)  PAC HTTP 代理服务器信息
    * **socks4**   - (OPTIONAL)  PAC SOCKET4 代理服务器信息
    * **socks5**   - (OPTIONAL) PAC SOCKS5 代理服务器信息
    * **local_rule_file**   - (OPTIONAL) 本地PAC规则文件 (一行填写一个域名)
    * **remote_rule_file**  - (OPTIONAL) 远程PAC规则文件 [bricks](https://raw.githubusercontent.com/Leask/BRICKS/master/gfw.bricks)

*  **proxies**             	- 代理配置项
	*  **http**       		- (OPTIONAL) 启用HTTP代理 (127.0.0.1:8080 / :8080)
	*  **socks4**          	- (OPTIONAL) 启用 SOCKS4 代理 (127.0.0.1:9090 / :9090)
	*  **socks5**          	- (OPTIONAL) 启用 SOCKS5 代理 (127.0.0.1:9999 / :9999)
	*  **crypto**   		- (OPTIONAL) SOCKS5的加密方法, 现在支持 rc4, des, aes-128-cfb, aes-192-cfb and aes-256-cfb
	*  **password**      	- 如果你设置了 **crypto**, 在这里就填写加密密码
	*  **dnsCacheTimeout**  - (OPTIONAL) 启用 dns 缓存 (单位为秒)
	*  **upstreams**		    - **upstream** 数组

* **upstream**
    *  **type**         	- 指定上游代理服务器的类型。现在支持shadowsocks和SOCKS5
    *  **crypto**        	- 指定上游代理服务器的加密方法。该加密方法同 **proxies.crypto**
    *  **password**         - 指定上游代理服务器的加密密码
    *  **address**          - 指定上游代理服务器的地址 (8.8.8.8:1111)
