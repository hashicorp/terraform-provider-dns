data "dns_recursive_cname_record_set" "hashicorp" {
  host = "www.hashicorp.com"
}

output "hashi_cnames" {
  value = data.dns_recursive_cname_record_set.hashicorp.cnames
}
