# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "dns_a_record_set" "google" {
  host = "google.com"
}

output "google_addrs" {
  value = join(",", data.dns_a_record_set.google.addrs)
}
