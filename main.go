package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-provider-dns/dns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: dns.Provider})
}
