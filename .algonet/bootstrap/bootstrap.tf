provider "aws" {
  region = "${var.state_region}"
}

resource "aws_s3_bucket" "b" {
  bucket = "${var.state_bucket}"  //tf-s3-algorand"
  acl    = "private"
  force_destroy = true
  versioning {
    enabled = true
  }
}

resource "aws_dynamodb_table" "terraform_statelock" {
  name           = "${var.state_table}"  //"algorand-state-lock"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}
