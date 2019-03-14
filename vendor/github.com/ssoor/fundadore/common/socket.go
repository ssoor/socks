package common

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

var (
	ErrorSocketUnavailable error = errors.New("socket port not find")
	ErrorNotValidAddress         = errors.New("Not a valid link address")
)

func GetConnectIP(connType string, connHost string) (ip string, err error) { //Get ip
	conn, err := net.Dial(connType, connHost)
	if err != nil {
		return ip, err
	}

	defer conn.Close()

	strSplit := strings.Split(conn.LocalAddr().String(), ":")

	if len(strSplit) < 2 {
		return ip, ErrorNotValidAddress
	}

	return strSplit[0], nil
}

func GetConnectMAC(connType string, connHost string) (string, error) {
	ip, err := GetConnectIP(connType, connHost)
	if nil != err {
		return "", err
	}

	interfaces, err := net.Interfaces() // 获取本机的MAC地址
	if err != nil {
		return "", err
	}

	if len(interfaces) == 0 {
		return "", errors.New("not find network hardware interface")
	}

	for _, inter := range interfaces {
		interAddrs, err := inter.Addrs()
		if nil != err {
			continue
		}

		for _, addr := range interAddrs {

			strSplit := strings.Split(addr.String(), "/")

			if len(strSplit) < 2 {
				continue
			}

			if strings.EqualFold(ip, strSplit[0]) {
				return inter.HardwareAddr.String(), nil
			}
		}
	}

	return "", errors.New("unknown error")
}

func GetIPs() (ips []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

func GetMACs() (MACs []string, err error) {
	interfaces, err := net.Interfaces() // 获取本机的MAC地址
	if err != nil {
		return nil, err
	}

	for _, inter := range interfaces {
		MACs = append(MACs, inter.HardwareAddr.String()) // 获取本机MAC地址
	}

	return MACs, nil
}

// 得到地址对应的端口.
func SocketGetPortFormAddr(addr string) (host string, port uint16, err error) {
	var intPort int
	var portString string

	if host, portString, err = net.SplitHostPort(addr); err != nil {
		return "", 0, err
	}

	if intPort, err = strconv.Atoi(portString); err != nil {
		return "", 0, err
	}

	port = uint16(intPort)
	return host, port, nil
}

// 得到一个可用的端口.
func SocketSelectPort(port_type string) (port uint16, err error) {
	listener, err := net.Listen(port_type, "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	_, port, err = SocketGetPortFormAddr(listener.Addr().String())

	return port, err
}

// 得到一个可用的端口.
func SocketSelectAddr(port_type string, ip string) (addr string, err error) {
	listener, err := net.Listen(port_type, ip+":0")
	if err != nil {
		return "", err
	}
	defer listener.Close()

	return listener.Addr().String(), nil
}
