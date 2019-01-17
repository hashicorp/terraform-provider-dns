---
layout: "dns"
page_title: "DNS: dns_mx_record_set"
sidebar_current: "docs-dns-mx-record-set"
description: |-
  Creates an MX type DNS record set.
---

# dns_mx_record_set

Creates an MX type DNS record set.

## Example Usage

```hcl
resource "dns_a_record_set" "smtp" {
  zone = "example.com."
  name = "smtp"
  ttl  = 300

  addresses = [
    "192.0.2.1",
  ]
}

resource "dns_a_record_set" "backup" {
  zone = "example.com."
  name = "backup"
  ttl  = 300

  addresses = [
    "192.0.2.2",
  ]
}

resource "dns_mx_record_set" "mx" {
  zone = "example.com."
  ttl  = 300

  mx {
    preference = 10
    exchange   = "smtp.example.com."
  }

  mx {
    preference = 20
    exchange   = "backup.example.com."
  }

  depends_on = [
    "dns_a_record_set.smtp",
    "dns_a_record_set.backup",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing dot.
* `name` - (Optional) The name of the record set. The `zone` argument will be appended to this value to create the full record path.
* `mx` - (Required) Can be specified multiple times for each MX record. Each block supports fields documented below.
* `ttl` - (Optional) The TTL of the record set. Defaults to `3600`.

The `mx` block supports:

* `preference` - (Required) The preference for the record.
* `exchange` - (Required) The FQDN of the mail exchange, include the trailing dot.

## Attributes Reference

The following attributes are exported:

* `zone` - See Argument Reference above.
* `name` - See Argument Reference above.
* `mx` - See Argument Reference above.
* `ttl` - See Argument Reference above.

## Import

Records can be imported using the FQDN, e.g.

```shell
$ terraform import dns_mx_record_set.mx example.com.
```
