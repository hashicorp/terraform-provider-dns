# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "dns_ptr_record_set" "hashicorp" {
  ip_address = "8.8.8.8"
}

output "hashi_ptr" {
  value = data.dns_ptr_record_set.hashicorp.ptr
}
