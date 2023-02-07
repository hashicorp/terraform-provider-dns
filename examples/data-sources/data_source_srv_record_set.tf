data "dns_srv_record_set" "sip" {
  service = "_sip._tcp.example.com."
}

output "sipserver" {
  value = data.dns_srv_record_set.sip.srv.0.target
}
