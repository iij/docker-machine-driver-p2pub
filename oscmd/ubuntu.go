package oscmd

import "fmt"

type Ubuntu struct {
	Linux
}

func (l Ubuntu) OpenFW(port int, proto string) []string {
	return []string{fmt.Sprintf("ufw allow %d/%s", port, proto)}
}

func (l Ubuntu) DefGW(addr string) []string {
	return []string{fmt.Sprintf("echo gateway %s | tee -a /etc/network/interfaces", addr)}
}

func (l Ubuntu) DNS(addrs []string) []string {
	res := []string{}
	for _, v := range addrs {
		res = append(res, fmt.Sprintf("echo nameserver %s | tee -a /etc/resolvconf/resolv.conf.d/base", v))
	}
	res = append(res, "resolvconf -u")
	return res
}
