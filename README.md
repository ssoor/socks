# SOCKS
[![Build Status](https://travis-ci.org/eahydra/socks.svg?branch=master)](https://travis-ci.org/eahydra/socks)  [![GoDoc](https://godoc.org/github.com/eahydra/socks?status.svg)](https://godoc.org/github.com/eahydra/socks)

SOCKS 实现了 SOCKS4/5 代理协议以及 HTTP 代理隧道, 你可以通过使用这些来简化你编写代理的难度.
 [cmd/socksd](https://github.com/eahydra/socks/blob/master/cmd/socksd) 这个项目是一个使用 socks 包构建的转换代理, 用于将 shadowsocks 或者 socsk5 转换成 SOCKS4/5 或者 HTTP 代理.
 
 目前 socks 支持的加密方式有 rc4, des, aes-128-cfb, aes-192-cfb , aes-256-cfb, 后端协议支持 shadowsocks 或者 socsk5.

# Install
如果你需要从源码安装, 可以尝试执行如下指令.
```
go get github.com/ssoor/socks
```

如果你想获得可运行文件, 请编译 [cmd/socksd](https://github.com/ssoor/socks/blob/master/cmd/socksd), 编译时可以参考此文档 [README.md](https://github.com/ssoor/socks/blob/master/cmd/socksd/README.md)