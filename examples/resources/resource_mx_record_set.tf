resource "dns_a_record_set" "smtp" {
  zone = "example.com."
  name = "smtp"
  ttl  = 300

  addresses = [
    "192.0.2.1",
  ]
}

resource "dns_a_record_set" "backup" {
  zone = "example.com."
  name = "backup"
  ttl  = 300

  addresses = [
    "192.0.2.2",
  ]
}

resource "dns_mx_record_set" "mx" {
  zone = "example.com."
  ttl  = 300

  mx {
    preference = 10
    exchange   = "smtp.example.com."
  }

  mx {
    preference = 20
    exchange   = "backup.example.com."
  }

  depends_on = [
    "dns_a_record_set.smtp",
    "dns_a_record_set.backup",
  ]
}
