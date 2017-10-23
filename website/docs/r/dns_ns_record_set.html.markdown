---
layout: "dns"
page_title: "DNS: dns_ns_record_set"
sidebar_current: "docs-dns-ns-record-set"
description: |-
  Creates a NA type DNS record set.
---

# dns_ns_record_set

Creates a NS type DNS record set.

## Example Usage

```hcl
resource "dns_ns_record_set" "www" {
  zone = "example.com."
  name = "www"
  nameservers = [
    "a.iana-servers.net.",
    "b.iana-servers.net.",
  ]
  ttl = 300
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing dot.
* `name` - (Required) The name of the record set. The `zone` argument will be appended to this value to create the full record path.
* `nameservers` - (Required) The nameservers this record set will point to.
* `ttl` - (Optional) The TTL of the record set. Defaults to `3600`.

## Attributes Reference

The following attributes are exported:

* `zone` - See Argument Reference above.
* `name` - See Argument Reference above.
* `nameservers` - See Argument Reference above.
* `ttl` - See Argument Reference above.
