resource "dns_txt_record_set" "google" {
  zone = "example.com."
  txt = [
    "google-site-verification=...",
  ]
  ttl = 300
}
