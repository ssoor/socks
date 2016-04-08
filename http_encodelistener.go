package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/ssoor/youniverse/log"
)

const MaxHeaderSize = 4
const MaxBufferSize = 0x1000
const MaxEncodeSize = uint16(0xFFFF)

// NewHTTPLPProxy constructs one HTTPLPProxy
func NewHTTPEncodeListener(l net.Listener) *LPListener {
	return &LPListener{listener: l}
}

type ECipherConn struct {
	net.Conn
	rwc io.ReadWriteCloser

	isPass   bool
	needRead []byte

	decodeSize int
	decodeCode byte
	decodeHead [MaxHeaderSize]byte
}

func (this *ECipherConn) getEncodeSize(encodeHeader []byte) (int, error) {

	if 0xCD != encodeHeader[0] {
		return 0, errors.New(fmt.Sprint("unrecognizable encryption header checksum: ", encodeHeader[0]))
	}

	if encodeHeader[3] != (encodeHeader[0] ^ (encodeHeader[1] + encodeHeader[2])) {
		return 0, errors.New(fmt.Sprint("encryption header information check fails: ", encodeHeader[3], ",Unexpected value: ", (encodeHeader[0] ^ encodeHeader[1] + encodeHeader[2])))
	}

	return int(binary.BigEndian.Uint16(encodeHeader[1:3])), nil
}

func (this *ECipherConn) Read(data []byte) (lenght int, err error) {

	if this.isPass { // 如果发生过错误 ,直接调用原始函数
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

	if 0 == len(this.needRead) { // 一个新的数据包
		this.isPass = true // 数据默认不需要解密，直接放过

		//log.Info("HTTP read data, need data size is ", len(data))
		if lenght, err = io.ReadFull(this.rwc, this.decodeHead[:MaxHeaderSize]); nil == err { // 检测数据包是否为加密包或者有效的 HTTP 包

			this.needRead = this.decodeHead[:MaxHeaderSize] // 数据需要发送

			if lenght, err = this.getEncodeSize(this.decodeHead[:MaxHeaderSize]); nil == err && lenght <= int(MaxEncodeSize) {
				this.decodeSize = lenght
				this.decodeCode = this.decodeHead[3]

				this.isPass = false // 数据需要解密
				this.needRead[0] = 'G'
				this.needRead[1] = 'E'
				this.needRead[2] = 'T'
				this.needRead[3] = ' '

				log.Infof("Socksd encode code: % 5d , encode len: %d\n", this.decodeCode, this.decodeSize)
			}
		}
	}

	if 0 != len(this.needRead) { // 发送缓冲区中的数据
		bufSize := len(data)

		if bufSize > len(this.needRead) {
			bufSize = len(this.needRead)
		}
		//bufSize:= min(len(data),len(this.needRead))

		read := this.needRead[:bufSize]
		this.needRead = this.needRead[bufSize:]

		//log.Info("Sending read data:", this.needRead, ", buffer size:", bufSize, ", data size:", len(this.needRead))

		return copy(data, read), nil
	}

	return this.rwc.Read(data)
}

func (c *ECipherConn) Write(data []byte) (int, error) {
	return c.rwc.Write(data)
}

func (c *ECipherConn) Close() error {
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

	return &ECipherConn{
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
