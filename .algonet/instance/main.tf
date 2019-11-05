provider "aws" {
  region = "${var.region}"
}

resource "aws_key_pair" "generated_key" {
  key_name   = "${var.region_key_pair_name}"
  public_key = "${var.public_key_ssh}"
}

data "aws_ami" "ami_name" {
  most_recent = true

  filter {
    name   = "name"
    values = ["${var.ami_filter_name}"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["${var.ami_filter_owner}"] # Canonical
}
resource "aws_instance" "node" {
  #ubuntu 18.04 LTS in us-east-1
  # ami = "${var.ami}"
  ami = "${data.aws_ami.ami_name.id}"
  instance_type = "${var.instance_type}"
  security_groups = "${var.security_groups}"
  associate_public_ip_address = true
  key_name      = "${aws_key_pair.generated_key.key_name}"
  root_block_device {
    volume_size = "${var.ebs_volume_size}"
    volume_type = "${var.ebs_volume_type}"
  }
  tags = {
    Name = "${var.ec2_name}"
    Role = "${var.ec2_role}"
    For = "${var.For}"
  }

  provisioner "file" {
    source      = "${var.script_path}"
    destination = "/tmp/${var.script}.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "sleep 50",
      "chmod +x /tmp/${var.script}.sh",
      "NETWORK=${var.network_name} CHANNEL=${var.channel_name} NODECFGHOST=${var.ec2_name} /tmp/${var.script}.sh ${var.region} ${var.instance_type} ${data.aws_ami.ami_name.id} ${var.ec2_name} ${var.region_key_pair_name} ${jsonencode(var.ports)} ${var.ebs_volume_size} ${var.ebs_volume_type} ${var.ebs_device_name}"
    ]
  }

  connection {
    type        = "ssh"
    private_key = "${var.private_key_pem}"
    user        = "ubuntu"
    timeout     = "3m"
    agent       = false
    host        = self.public_ip
  }
}
