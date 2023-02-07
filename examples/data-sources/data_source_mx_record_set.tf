data "dns_mx_record_set" "mail" {
  domain = "example.com."
}

output "mailserver" {
  value = data.dns_mx_record_set.mail.mx.0.exchange
}
