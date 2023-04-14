resource "dns_aaaa_record_set" "www" {
  zone = "example.com."
  name = "www"
  addresses = [
    "fdd5:e282:43b8:5303:dead:beef:cafe:babe",
    "fdd5:e282:43b8:5303:cafe:babe:dead:beef",
  ]
  ttl = 300
}
