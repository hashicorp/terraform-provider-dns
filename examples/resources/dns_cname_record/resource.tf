# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "dns_cname_record" "foo" {
  zone  = "example.com."
  name  = "foo"
  cname = "bar.example.com."
  ttl   = 300
}
