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
