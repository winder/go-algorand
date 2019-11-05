variable "region" {}
variable "instance_type" {}
variable "ec2_name" {}
variable "ec2_role" {}
variable "For" {}
variable "region_key_pair_name" {}
variable "ebs_device_name" {}
variable "ebs_volume_type" {}
variable "ebs_volume_size" {}
variable "script" {}
variable "script_path" {}
variable "ports" {
  type = "list"
}
variable "security_groups" {
  type = "list"
}
variable "private_key_pem" {}
variable "public_key_pem" {}
variable "public_key_ssh" {}

variable "ami_filter_name" {}
variable "ami_filter_owner" {}
variable "network_name" {}
variable "channel_name" {}
