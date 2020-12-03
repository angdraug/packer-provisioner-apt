package provisioner

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	config Config
	comm   packer.Communicator
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	return p.config.Prepare(raws...)
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, _ map[string]interface{}) error {
	ui.Say("Provisioning with APT...")

	err := comm.UploadDir("/var/cache/apt/archives", p.config.CacheDir, []string{})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to upload APT cache from %s", p.config.CacheDir))
		return err
	}

	cmd := &packer.RemoteCmd{
		Command: "/bin/sh -c 'for i in $(seq 100); do " +
			"resolvectl query deb.debian.org >/dev/null && break; sleep 0.1; done; " +
			"resolvectl query deb.debian.org'",
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		ui.Error("Failed waiting for domain name resolution")
		return err
	}

	for _, key := range p.config.Keys {
		f, err := os.Open(key)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		err = comm.Upload("/etc/apt/trusted.gpg.d/"+filepath.Base(key), f, &fi)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to upload APT key %s", key))
			return err
		}
	}

	if len(p.config.Sources) != 0 {
		r := strings.NewReader(strings.Join(p.config.Sources, "\n") + "\n")
		err := comm.Upload("/etc/apt/sources.list.d/packer.list", r, nil)
		if err != nil {
			ui.Error("Failed to upload APT sources list")
			return err
		}
		cmd := &packer.RemoteCmd{Command: "/usr/bin/apt-get update"}
		if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
			ui.Error("apt-get update failed")
			return err
		}
	}

	cmd = &packer.RemoteCmd{
		Command: fmt.Sprintf(
			"DEBIAN_FRONTEND=noninteractive /usr/bin/apt-get install -y --no-install-recommends %s",
			strings.Join(p.config.Packages, " "),
		),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		ui.Error("apt-get install failed.")
		return err
	}

	p.updateCache(ui, comm)

	cmd = &packer.RemoteCmd{Command: "/usr/bin/apt-get clean"}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		ui.Error("apt-get clean failed, ignoring")
	}

	return nil
}

func (p *Provisioner) updateCache(ui packer.Ui, comm packer.Communicator) {
	dir, err := ioutil.TempDir(os.TempDir(), "archives-")
	if err != nil {
		ui.Error("APT cache update: failed to create tempdir")
		return
	}
	defer os.RemoveAll(dir)

	if err := comm.DownloadDir("/var/cache/apt/archives", dir, []string{}); err != nil {
		ui.Error(fmt.Sprintf("APT cache update: failed to download archives to %s", dir))
		return
	}

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("mv -n %s/*.deb %s", dir, p.config.CacheDir))
	if err := cmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("APT cache update: mv: %v", err))
	}
}
