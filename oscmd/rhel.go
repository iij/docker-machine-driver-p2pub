package oscmd

import "fmt"

type RedHat struct {
	Linux
}

func (l RedHat) OpenFW(port int, proto string) []string {
	return []string{
		fmt.Sprintf("firewall-cmd --add-port %d/%s", port, proto),
		fmt.Sprintf("firewall-cmd --add-port %d/%s --permanent", port, proto),
	}
}

func (l RedHat) DefGW(addr string) []string {
	return []string{
		fmt.Sprintf("nmcli con mod eth0 ipv4.gateway %s", addr),
		fmt.Sprintf("ip route add default via %s", addr),
	}
}

func (l RedHat) ARP() []string {
	return []string{
		"echo 'net.ipv4.conf.default.arp_ignore = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_announce = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_notify = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_filter = 0' >> /etc/sysctl.conf",
		"echo 'net.ipv4.conf.default.arp_accept = 0' >> /etc/sysctl.conf",
		"sysctl -p",
	}
}

type CentOS struct {
	RedHat
}
