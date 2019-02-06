---
layout: "dns"
page_title: "DNS: dns_txt_record_set"
sidebar_current: "docs-dns-datasource-txt-record-set"
description: |-
  Get DNS TXT record set.
---

# dns_txt_record_set

Use this data source to get DNS TXT record set of the host.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that the list of `records` is not empty. 

## Example Usage

```hcl
data "dns_txt_record_set" "hashicorp" {
  host = "www.hashicorp.com"
}

output "hashi_txt" {
  value = "${data.dns_txt_record_set.hashi.record}"
}

output "hashi_txts" {
  value = "${join(",", data.dns_txt_record_set.hashi.records)}"
}
```

## Argument Reference

The following arguments are supported:

 * `host` - (required): Host to look up
 
 * `ignore_errors` - (optional, default: `false`): When `true` and the DNS record cannot be resolved, 
   return an empty list in `records` and an empty string in `record`.

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `host`.

 * `record` - The first TXT record.

 * `records` - A list of TXT records.
