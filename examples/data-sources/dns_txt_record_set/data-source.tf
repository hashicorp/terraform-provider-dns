data "dns_txt_record_set" "hashicorp" {
  host = "www.hashicorp.com"
}

output "hashi_txt" {
  value = data.dns_txt_record_set.hashicorp.record
}

output "hashi_txts" {
  value = join(",", data.dns_txt_record_set.hashicorp.records)
}
