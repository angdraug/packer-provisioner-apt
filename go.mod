module git.sr.ht/~angdraug/packer-provisioner-apt

go 1.15

require (
	github.com/hashicorp/hcl/v2 v2.7.0
	github.com/hashicorp/packer v1.6.4
	github.com/zclconf/go-cty v1.5.0
)

replace github.com/hashicorp/packer => github.com/angdraug/packer v1.6.4-ugorji-go-v1.1.13
