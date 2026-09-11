data "alicloud_account" "current" {}

resource "alicloud_vpc" "main" {
  vpc_name   = var.name
  cidr_block = var.vpc_cidr
}

resource "alicloud_vswitch" "main" {
  vpc_id       = alicloud_vpc.main.id
  cidr_block   = cidrsubnet(var.vpc_cidr, 8, 0)
  zone_id      = var.zone_id
  vswitch_name = var.name
}
