---
layout: "dns"
page_title: "DNS: dns_cname_record_set"
sidebar_current: "docs-dns-datasource-cname-record-set"
description: |-
  Get DNS CNAME record set.
---

# dns_cname_record_set

Use this data source to get DNS CNAME record set of the host.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that `cname` is not empty. 

## Example Usage

```hcl
data "dns_cname_record_set" "hashicorp" {
  host = "www.hashicorp.com"
}

output "hashi_cname" {
  value = "${data.dns_cname_record_set.hashicorp.cname}"
}
```

## Argument Reference

The following arguments are supported:

 * `host` - (required): Host to look up
 
 * `ignore_errors` - (optional, default: `false`): When `true` and the DNS record cannot be resolved, 
   return an empty string in `cname`.

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `host`.

 * `cname` - A CNAME record associated with host.
