output "vpc_id" {
  value = alicloud_vpc.main.id
}

output "account_id" {
  value = data.alicloud_account.current.id
}
