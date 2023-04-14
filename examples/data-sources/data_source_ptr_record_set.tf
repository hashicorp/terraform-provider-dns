data "dns_ptr_record_set" "hashicorp" {
  ip_address = "8.8.8.8"
}

output "hashi_ptr" {
  value = data.dns_ptr_record_set.hashicorp.ptr
}
