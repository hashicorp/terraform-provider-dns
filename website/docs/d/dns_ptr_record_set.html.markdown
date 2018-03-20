---
layout: "dns"
page_title: "DNS: dns_ptr_record_set"
sidebar_current: "docs-dns-datasource-ptr-record-set"
description: |-
  Get DNS PTR record set.
---

# dns_ptr_record_set

Use this data source to get DNS PTR record set of the ip address.

## Example Usage

```hcl
data "dns_ptr_record_set" "hashicorp" {
  ip_address = "8.8.8.8"
}

output "hashi_ptr" {
  value = "${data.dns_ptr_record_set.hashicorp.ptr}"
}
```

## Argument Reference

The following arguments are supported:

 * `ip_address` - (required): IP address to look up

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `ip_address`.

 * `ptr` - A PTR record associated with `ip_address`.

 __NOTE__: Only the first result is taken from the query.
