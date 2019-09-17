---
layout: "dns"
page_title: "DNS: dns_srv_record_set"
sidebar_current: "docs-dns-datasource-srv-record-set"
description: |-
  Get DNS SRV record set.
---

# dns_srv_record_set

Use this data source to get DNS SRV records for a service.

By default, querying for a non-existent record will result in an error and the plan will be aborted.
Conditional logic can be implemented by setting `ignore_errors` to `true` and checking
that the list of `srv` records is not empty. 

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

 * `service` - (required): Service to look up
 
 * `ignore_errors` - (optional, default: `false`): When `true` and the DNS record cannot be resolved, 
   return an empty list in `srv`.

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `service`.
 
 * `srv` - A list of records. They are sorted to stay consistent across runs.

The `srv` block supports:

* `priority` - The priority of the target host.

* `weight` - A relative weight for records with the same priority.

* `port` - The TCP or UDP port on which the service is to be found.

* `target` - The canonical hostname of the machine providing the service, ending in a dot.
