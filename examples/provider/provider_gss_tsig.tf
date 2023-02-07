# Configure the DNS Provider
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