---
layout: "dns"
page_title: "DNS: dns_aaaa_record_set"
description: |-
  Get DNS AAAA record set.
---

# dns_aaaa_record_set

Use this data source to get DNS AAAA records of the host.

## Example Usage

```hcl
data "dns_aaaa_record_set" "google" {
  host = "google.com"
}

output "google_addrs" {
  value = join(",", data.dns_aaaa_record_set.google.addrs)
}
```

## Argument Reference

The following arguments are supported:

 * `host` - (required): Host to look up

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `host`.

 * `addrs` - A list of IP addresses. IP addresses are always sorted to avoid constant changing plans.
