---
layout: "dns"
page_title: "DNS: dns_srv_record_set"
sidebar_current: "docs-dns-datasource-srv-record-set"
description: |-
  Get DNS SRV record set.
---

# dns_srv_record_set

Use this data source to get DNS SRV records for a service.

## Example Usage

```hcl
data "dns_srv_record_set" "sip" {
  service = "_sip._tcp.example.com."
}

output "sipserver" {
  value = "${data.dns_srv_record_set.sip.srv.0.target}"
}
```

## Argument Reference

The following arguments are supported:

 * `service` - (Required): Service to look up

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `service`.
 * `srv` - A list of records. They are sorted to stay consistent across runs.
