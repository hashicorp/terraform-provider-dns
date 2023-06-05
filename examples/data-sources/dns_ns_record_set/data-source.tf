# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "dns_ns_record_set" "google" {
  host = "google.com"
}

output "google_nameservers" {
  value = join(",", data.dns_ns_record_set.google.nameservers)
}
