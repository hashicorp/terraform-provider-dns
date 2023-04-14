data "dns_aaaa_record_set" "google" {
  host = "google.com"
}

output "google_addrs" {
  value = join(",", data.dns_aaaa_record_set.google.addrs)
}
