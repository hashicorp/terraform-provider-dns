# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "dns_ptr_record" "dns-sd" {
  zone = "example.com."
  name = "r._dns-sd"
  ptr  = "example.com."
  ttl  = 300
}
