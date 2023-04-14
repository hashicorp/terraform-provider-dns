resource "dns_srv_record_set" "sip" {
  zone = "example.com."
  name = "_sip._tcp"
  srv {
    priority = 10
    weight   = 60
    target   = "bigbox.example.com."
    port     = 5060
  }
  srv {
    priority = 10
    weight   = 20
    target   = "smallbox1.example.com."
    port     = 5060
  }
  srv {
    priority = 10
    weight   = 20
    target   = "smallbox2.example.com."
    port     = 5060
  }
  ttl = 300
}
