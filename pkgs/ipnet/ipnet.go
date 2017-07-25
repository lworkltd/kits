package ipnet

import (
	"net"
)

// 获取本机IP，如果指定了网卡，则返回网卡的ipv4，否则选择第一个网卡的ip
func Ipv4(adapter string) (string, bool, error) {
	if adapter == "" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return "", false, err
		}
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
				if ipv4 := ip.IP.To4(); ipv4 != nil {
					return ipv4.String(), true, nil
				}
			}
		}
	}

	ifc, err := net.InterfaceByName(adapter)
	if err != nil {
		return "", false, err
	}

	addrs, err := ifc.Addrs()
	if err != nil {
		return "", false, err
	}

	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ipv4 := ip.IP.To4(); ipv4 != nil {
				return ipv4.String(), true, nil
			}
		}
	}

	return "", true, nil
}
