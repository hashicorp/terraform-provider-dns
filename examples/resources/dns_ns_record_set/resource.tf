resource "dns_ns_record_set" "www" {
  zone = "example.com."
  name = "www"
  nameservers = [
    "a.iana-servers.net.",
    "b.iana-servers.net.",
  ]
  ttl = 300
}
