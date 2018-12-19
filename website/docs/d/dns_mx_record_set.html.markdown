---
layout: "dns"
page_title: "DNS: dns_mx_record_set"
sidebar_current: "docs-dns-datasource-mx-record-set"
description: |-
  Get DNS mx record set.
---

# dns_mx_record_set

Use this data source to get DNS mx records of the zone.

## Example Usage

```hcl
data "dns_mx_record_set" "google" {
  zone = "google.com"
}

output "google_mxservers" {
  value = "${join(",", data.dns_mx_record_set.google.mxservers)}"
}

output "google_mxserver_priorities" {
  value = "${join(",", data.dns_mx_record_set.google.priorities)}"
}
```

## Argument Reference

The following arguments are supported:

 * `zone` - (required): Zone to look up

## Attributes Reference

The following attributes are exported:

 * `id` - Set to `zone`.

 * `mxservers` - A list of MX servers. MX servers are sorted first by priority and then by name to avoid changing plans.

 * `priorities` - A list of MX server priorities corresponding to the mxserver list entries.
