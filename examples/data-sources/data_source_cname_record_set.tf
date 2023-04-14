data "dns_cname_record_set" "hashicorp" {
  host = "www.hashicorp.com"
}

output "hashi_cname" {
  value = data.dns_cname_record_set.hashicorp.cname
}
