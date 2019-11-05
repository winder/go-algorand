provider "aws" {
  region = "${var.region}"
}

resource "aws_security_group" "node" {
  count = "${length(var.ports)}"
  name = "${var.ports[count.index]}_${var.network_name}"
  description = "grant ssh"

  ingress {
    from_port   = "${var.ports[count.index]}"
    to_port     = "${var.ports[count.index]}"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
