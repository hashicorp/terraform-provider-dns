---
layout: "dns"
page_title: "DNS: dns_srv_record_set"
description: |-
  Creates an SRV type DNS record set.
---

# dns_srv_record_set

Creates an SRV type DNS record set.

## Example Usage

```hcl
resource "dns_srv_record_set" "sip" {
  zone = "example.com."
  name = "_sip._tcp"
  srv {
    priority = 10
    weight   = 60
    target   = "bigbox.example.com."
    port     = 5060
  }
  srv {
    priority = 10
    weight   = 20
    target   = "smallbox1.example.com."
    port     = 5060
  }
  srv {
    priority = 10
    weight   = 20
    target   = "smallbox2.example.com."
    port     = 5060
  }
  ttl = 300
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing dot.
* `name` - (Required) The name of the record set. The `zone` argument will be appended to this value to create the full record path.
* `srv` - (Required) Can be specified multiple times for each SRV record. Each block supports fields documented below.
* `ttl` - (Optional) The TTL of the record set. Defaults to `3600`.

The `srv` block supports:

* `priority` - (Required) The priority for the record.
* `weight` - (Required) The weight for the record.
* `target` - (Required) The FQDN of the target, include the trailing dot.
* `port` - (Required) The port for the service on the target.

## Attributes Reference

The following attributes are exported:

* `zone` - See Argument Reference above.
* `name` - See Argument Reference above.
* `srv` - See Argument Reference above.
* `ttl` - See Argument Reference above.

## Import

Records can be imported using the FQDN, e.g.

```
$ terraform import dns_srv_record_set.sip _sip._tcp.example.com.
```
