---
layout: "dns"
page_title: "DNS: dns_a_record_set"
sidebar_current: "docs-dns-datasource-a-record-set"
description: |-
  Get DNS A record set.
---

# dns_a_record_set

Use this data source to get DNS A records of the host.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that the list of `addrs` is not empty. 

## Example Usage

```hcl
data "dns_a_record_set" "google" {
  host = "google.com"
}

output "google_addrs" {
  value = "${join(",", data.dns_a_record_set.google.addrs)}"
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
