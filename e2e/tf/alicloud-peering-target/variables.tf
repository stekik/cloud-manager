variable "region" {
  description = "AliCloud region for the remote VPC"
  type        = string
  default     = "ap-northeast-1"
}

variable "name" {
  description = "Name prefix for all resources"
  type        = string
  nullable    = false
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "192.168.0.0/16"
}

variable "zone_id" {
  description = "Availability zone for the vswitch. Hardcoded to avoid DescribeZones API flakiness."
  type        = string
  default     = "ap-northeast-1a"
}
