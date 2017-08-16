package oscmd

import "fmt"

type Oscmd interface {
	OpenFW(port int, proto string) []string
	DefGW(addr string) []string
	DNS(addrs []string) []string
	ARP() []string
}

type Linux struct{}

func (l Linux) DefGW(addr string) []string {
	return []string{fmt.Sprintf("ip route add default via %s", addr)}
}

func (l Linux) DNS(addrs []string) []string {
	res := []string{}
	for _, v := range addrs {
		res = append(res, fmt.Sprintf("echo nameserver %s | tee -a /etc/resolv.conf", v))
	}
	return res
}

func (l Linux) ARP() []string {
	return []string{
		"echo 'net.ipv4.conf.default.arp_ignore = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_announce = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_notify = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_filter = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_accept = 0' >> /etc/sysctl.conf",
		"sysctl -p",
	}
}
