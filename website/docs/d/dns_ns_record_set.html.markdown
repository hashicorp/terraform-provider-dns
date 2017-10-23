---
layout: "dns"
page_title: "DNS: dns_ns_record_set"
sidebar_current: "docs-dns-datasource-ns-record-set"
description: |-
  Get DNS ns record set.
---

# dns_ns_record_set

Use this data source to get DNS ns records of the host.

## Example Usage

```hcl
data "dns_ns_record_set" "google" {
  host = "google.com"
}

output "google_nameservers" {
  value = "${join(",", data.dns_ns_record_set.google.nameservers)}"
}
```

## Argument Reference

The following arguments are supported:

 * `host` - (required): Host to look up

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `host`.

 * `nameservers` - A list of nameservers. Nameservers are always sorted to avoid constant changing plans.
