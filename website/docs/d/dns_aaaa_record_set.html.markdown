---
layout: "dns"
page_title: "DNS: dns_aaaa_record_set"
sidebar_current: "docs-dns-datasource-aaaa-record-set"
description: |-
  Get DNS AAAA record set.
---

# dns_aaaa_record_set

Use this data source to get DNS AAAA records of the host.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that the list of `addrs` is not empty. 

## Example Usage

```hcl
data "dns_aaaa_record_set" "google" {
  host = "google.com"
}

output "google_addrs" {
  value = "${join(",", data.dns_aaaa_record_set.google.addrs)}"
}
```

## Argument Reference

The following arguments are supported:

 * `host` - (required): Host to look up
 
 * `ignore_errors` - (optional, default: `false`): When `true` and the DNS record cannot be resolved, 
   return an empty list in `addrs`.

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `host`.

 * `addrs` - A list of IP addresses. IP addresses are always sorted to avoid constant changing plans.
