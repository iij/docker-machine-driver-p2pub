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

type CentOS struct {
	RedHat
}
