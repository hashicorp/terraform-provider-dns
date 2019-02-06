---
layout: "dns"
page_title: "DNS: dns_ns_record_set"
sidebar_current: "docs-dns-datasource-ns-record-set"
description: |-
  Get DNS ns record set.
---

# dns_ns_record_set

Use this data source to get DNS NS records of the host.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that the list of `nameservers` is not empty. 

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
 
 * `ignore_errors` - (optional, default: `false`): When `true` and the DNS record cannot be resolved, 
   return an empty list in `nameservers`.

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `host`.

 * `nameservers` - A list of nameservers. Nameservers are always sorted to avoid constant changing plans.
