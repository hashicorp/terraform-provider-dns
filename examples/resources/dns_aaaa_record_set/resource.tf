resource "dns_aaaa_record_set" "www" {
  zone = "example.com."
  name = "www"
  addresses = [
    "fdd5:e282:43b8:5303:1234:5678:cafe:9012",
    "fdd5:e282:43b8:5303:cafe:9012:1234:5678",
  ]
  ttl = 300
}
