package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// NewHTTPLPProxy constructs one HTTPLPProxy
func NewHTTPLPListener(l net.Listener) *LPListener {
	return &LPListener{listener: l}
}

type CipherConn struct {
	net.Conn
	iserror bool
	rwc     io.ReadWriteCloser

	decodeSize int
	decodeCode byte
}

func (this *CipherConn) getEncodeSize(encodeHeader []byte) (int, error) {

	if 0xCD != encodeHeader[0] {
		return 0, errors.New(fmt.Sprint("unrecognizable encryption header checksum: ", encodeHeader[0]))
	}

	if encodeHeader[3] != (encodeHeader[0] ^ (encodeHeader[1] + encodeHeader[2])) {
		return 0, errors.New(fmt.Sprint("encryption header information check fails: ", encodeHeader[3], ",Unexpected value: ", (encodeHeader[0] ^ encodeHeader[1] + encodeHeader[2])))
	}

	return int(binary.BigEndian.Uint16(encodeHeader[1:3])), nil
}

const MaxBufferSize = 0x1000
const MaxEncodeSize = uint16(0xFFFF)

func (this *CipherConn) Read(data []byte) (lenght int, err error) {

	if this.iserror { // 如果发生过错误 ,直接调用原始函数
		return this.rwc.Read(data)
	}

	if 0 != this.decodeSize { //解密头已获得,进入解密流程
		if lenght, err = this.rwc.Read(data); err != nil {
			return
		}

		if lenght > this.decodeSize {
			lenght = this.decodeSize
		}

		for i := 0; i < int(lenght); i++ {
			data[i] ^= this.decodeCode | 0x80
		}

		this.decodeSize -= lenght
		//fmt.Print(string(data[:lenght]))

		return
	}

	this.iserror = true // 默认解密失败

	if lenght, err = io.ReadFull(this.rwc, data[:4]); nil != err { //如果接收不够4字节说明这不是一个有效的HTTP包或者加密包
		return
	}

	//fmt.Println("Read Data: ", string(data[:4]))

	lenght, err = this.getEncodeSize(data[:4])

	if nil == err && lenght <= int(MaxEncodeSize) {
		this.decodeSize = lenght
		this.decodeCode = data[3]

		fmt.Println("Encode Code: ", this.decodeCode, ",Encode Len:", this.decodeSize)

		data[0] = 'G'
		data[1] = 'E'
		data[2] = 'T'
		data[3] = ' '

		this.iserror = false //修改状态为解密正常
	}

	return 4, nil

}

func (c *CipherConn) Write(data []byte) (int, error) {
	return c.rwc.Write(data)
}

func (c *CipherConn) Close() error {
	err := c.Conn.Close()
	c.rwc.Close()
	return err
}

type LPListener struct {
	listener net.Listener
}

func (this *LPListener) Accept() (c net.Conn, err error) {
	conn, err := this.listener.Accept()

	if err != nil {
		return nil, err
	}

	return &CipherConn{
		Conn: conn,
		rwc:  conn,
	}, nil
}

func (this *LPListener) Close() error {
	return this.listener.Close()
}

func (this *LPListener) Addr() net.Addr {
	return this.listener.Addr()
}
