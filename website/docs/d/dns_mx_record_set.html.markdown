---
layout: "dns"
page_title: "DNS: dns_mx_record_set"
sidebar_current: "docs-dns-datasource-mx-record-set"
description: |-
  Get DNS MX record set.
---

# dns_mx_record_set

Use this data source to get DNS MX records for a domain.

## Example Usage

```hcl
data "dns_mx_record_set" "mail" {
  domain = "example.com."
}

output "mailserver" {
  value = data.dns_mx_record_set.mail.mx.0.exchange
}
```

## Argument Reference

The following arguments are supported:

 * `domain` - (Required): Domain to look up

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `service`.
 * `mx` - A list of records. They are sorted by ascending preference then alphabetically by exchange to stay consistent across runs.
