---
layout: "dns"
page_title: "DNS: dns_txt_record_set"
description: |-
  Creates a TXT type DNS record set.
---

# dns_txt_record_set

Creates a TXT type DNS record set.

## Example Usage

```hcl
resource "dns_txt_record_set" "google" {
  zone = "example.com."
  txt = [
    "google-site-verification=...",
  ]
  ttl = 300
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing dot.
* `name` - (Optional) The name of the record set. The `zone` argument will be appended to this value to create the full record path.
* `txt` - (Required) The text records this record set will be set to.
* `ttl` - (Optional) The TTL of the record set. Defaults to `3600`.

## Attributes Reference

The following attributes are exported:

* `zone` - See Argument Reference above.
* `name` - See Argument Reference above.
* `txt` - See Argument Reference above.
* `ttl` - See Argument Reference above.

## Import

Records can be imported using the FQDN, e.g.

```
$ terraform import dns_txt_record_set.google example.com.
```
