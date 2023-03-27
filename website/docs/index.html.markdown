---
layout: "dns"
page_title: "Provider: DNS"
description: |-
  The DNS provider supports DNS updates (RFC 2136). Additionally, the provider can be configured with secret key based transaction authentication (RFC 2845) or can use GSS-TSIG (RFC 3645).
---

# DNS Provider

The DNS provider supports resources that perform DNS updates ([RFC 2136](https://datatracker.ietf.org/doc/html/rfc2136)) and data sources for reading DNS information. The provider can be configured with secret key based transaction authentication ([RFC 2845](https://datatracker.ietf.org/doc/html/rfc2845)) or GSS-TSIG ([RFC 3645](https://datatracker.ietf.org/doc/html/rfc3645)).

Use the navigation to the left to read about the available resources and data sources.

## Example Usage

Using secret key based transaction authentication (RFC 2845):

```hcl
# Configure the DNS Provider
provider "dns" {
  update {
    server        = "192.168.0.1"
    key_name      = "example.com."
    key_algorithm = "hmac-md5"
    key_secret    = "3VwZXJzZWNyZXQ="
  }
}

# Create a DNS A record set
resource "dns_a_record_set" "www" {
  # ...
}
```

Using GSS-TSIG (RFC 3645):

```hcl
provider "dns" {
  update {
    server = "ns.example.com" # Using the hostname is important in order for an SPN to match
    gssapi {
      realm    = "EXAMPLE.COM"
      username = "user"
      keytab   = "/path/to/keytab"
    }
  }
}

# Create a DNS A record set
resource "dns_a_record_set" "www" {
  # ...
}
```

## Configuration Reference

`update` - (Optional) When the provider is used for DNS updates, this block is required. Structure is documented below.

The `update` block supports the following attributes:

* `server` - (Required) The hostname or IP address of the DNS server to send updates to.
* `port` - (Optional) The target UDP port on the server where updates are sent to. Defaults to `53`.
* `transport` - (Optional) Transport to use for DNS queries. Valid values are `udp`, `udp4`, `udp6`, `tcp`, `tcp4`, or `tcp6`. Any UDP transport will retry automatically with the equivalent TCP transport in the event of a truncated response. Defaults to `udp`.
* `timeout` - (Optional) Timeout for DNS queries. Valid values are durations expressed as `500ms`, etc. or a plain number which is treated as whole seconds.
* `retries` - (Optional) How many times to retry on connection timeout. Defaults to `3`.
* `backoff` - (Optional) How many milliseconds the provider should back off requerying the DNS server for validation after pushing record updates. Should only be used in load balanced or otherwise slowly propagated environments. Defaults to `0`.
* `key_name` - (Optional) The name of the TSIG key used to sign the DNS update messages.
* `key_algorithm` - (Optional; Required if `key_name` is set) When using TSIG authentication, the algorithm to use for HMAC. Valid values are `hmac-md5`, `hmac-sha1`, `hmac-sha256` or `hmac-sha512`.
* `key_secret` - (Optional; Required if `key_name` is set)
    A Base64-encoded string containing the shared secret to be used for TSIG.
* `gssapi` - (Optional) A `gssapi` block (documented below). Only one `gssapi` block may be in the configuration. Conflicts with use of `key_name`, `key_algorithm` and `key_secret`.

### gssapi Configuration Block

The `gssapi` configuration block supports the following arguments:

* `realm` - (Required) The Kerberos realm or Active Directory domain.
* `username` - (Optional) The name of the user to authenticate as. If not set the current user session will be used.
* `password` - (Optional; This or `keytab` is required if `username` is set) The matching password for `username`.
* `keytab` - (Optional; This or `password` is required if `username` is set, not supported on Windows) The path to a keytab file containing a key for `username`.
