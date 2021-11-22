---
layout: "dns"
page_title: "DNS: dns_ptr_record_set"
sidebar_current: "docs-dns-ptr-record-set"
description: |-
  Creates a PTR type DNS record set.
---

# dns_ptr_record_set

Creates a PTR type DNS record set.

## Example Usage

```hcl
resource "dns_ptr_record_set" "dns-sd" {
  zone = "example.com."
  name = "r._dns-sd"
  ptrs  = ["example.com."]
  ttl  = 300
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) DNS zone the record belongs to. It must be an FQDN, that is, include the trailing dot.
* `name` - (Optional) The name of the record. The `zone` argument will be appended to this value to create the full record path.
* `ptrs` - (Required) A list of names that this record will point to.
* `ttl` - (Optional) The TTL of the record set. Defaults to `3600`.

## Attributes Reference

The following attributes are exported:

* `zone` - See Argument Reference above.
* `name` - See Argument Reference above.
* `ptrs` - See Argument Reference above.
* `ttl` - See Argument Reference above.

## Import

Records can be imported using the FQDN, e.g.

```
$ terraform import dns_ptr_record.dns-sd r._dns-sd.example.com.
```
