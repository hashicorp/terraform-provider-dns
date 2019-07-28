---
layout: "dns"
page_title: "DNS: dns_mx_record_set"
sidebar_current: "docs-dns-datasource-mx-record-set"
description: |-
  Get DNS MX record set.
---

# dns_mx_record_set

Use this data source to get DNS MX records for a domain.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that the list of `mx` records is not empty. 

## Example Usage

```hcl
data "dns_mx_record_set" "mail" {
  domain = "example.com."
}

output "mailserver" {
  value = "${data.dns_mx_record_set.mail.mx.0.exchange}"
}
```

## Argument Reference

The following arguments are supported:

 * `domain` - (required): Domain to look up
 
 * `ignore_errors` - (optional, default: `false`): When `true` and the DNS record cannot be resolved, 
   return an empty list in `mx`.

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `domain`.

 * `mx` - A list of records. They are sorted by ascending `preference` then alphabetically by `exchange` to stay consistent across runs.

The `mx` block supports:

* `preference` - Preference (priority) value for the MX record.

* `exchange` - Domain name of the mailserver.
