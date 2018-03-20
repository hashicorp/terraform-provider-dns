## 2.0.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:
* Prior versions of the provider would sign requests when sending updates to a DNS server but would not sign the requests to read those values back on subsequent refreshes. For consistency, now _read_ requests are also signed for managed resources in this provider. This does not apply to the data sources, which continue to just send normal unsigned DNS requests as before.

NEW FEATURES:
* Use signed requests when refreshing managed resources [GH-35]
* data/dns_ptr_record_set: Implement data source for PTR record. [GH-32]

BUGS FIXED:

* Normalize IP addresses before comparing them, so non-canonical forms don't cause errant diffs [GH-13]
* Properly handle IPv6 IP addresses as the update host. Previously this would create an invalid connection address due to not properly constructing the address format. [GH-22]
* When refreshing DNS record resources, `NXDOMAIN` errors are now properly marked as deletions in state rather than returning an error, thus allowing Terraform to plan to re-create the missing records. [GH-33]

## 1.0.0 (September 15, 2017)

* No changes from 0.1.1; just adjusting to [the new version numbering scheme](https://www.hashicorp.com/blog/hashicorp-terraform-provider-versioning/).

## 0.1.1 (August 28, 2017)

NEW FEATURES:

* **`dns_aaaa_record_set` data source** for fetching IPv6 address records ([#9](https://github.com/terraform-providers/terraform-provider-dns/issues/9))
* **`dns_ns_record_set` data source** for fetching nameserver records ([#10](https://github.com/terraform-providers/terraform-provider-dns/issues/10))
* **`dns_ns_record_set` resource** for creating new nameserver records via the DNS update protocol ([#10](https://github.com/terraform-providers/terraform-provider-dns/issues/10))

## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
