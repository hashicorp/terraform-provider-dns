package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-dns/dns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: dns.Provider})
}
