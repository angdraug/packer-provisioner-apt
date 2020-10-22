package main

import (
	"git.sr.ht/~angdraug/packer-provisioner-apt/provisioner"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(provisioner.Provisioner))
	server.Serve()
}
