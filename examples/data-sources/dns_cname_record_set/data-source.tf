# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "dns_cname_record_set" "hashicorp" {
  host = "www.hashicorp.com"
}

output "hashi_cname" {
  value = data.dns_cname_record_set.hashicorp.cname
}
