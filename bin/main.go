package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/iij/docker-machine-driver-p2pub"
)

func main() {
	plugin.RegisterDriver(new(p2pubmachine.Driver))
}
