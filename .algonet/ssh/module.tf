  resource "null_resource" "security" {

    provisioner "local-exec" {
      command = "rm -rf tf_private_key.pem"
    }
    provisioner "local-exec" {
      command = "echo '${var.public_key_pem}${var.private_key_pem}' >> tf_private_key.pem"
    }
    provisioner "local-exec" {
      command = "chmod 400 tf_private_key.pem"
    }
  }
