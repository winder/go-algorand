variable state_region {
  description = "Region where S3 bucket and DynamoDB table will be placed"
}
variable state_table {
  description = "DynamoDB table name, where locking state will be placed"
}
variable state_bucket {
  description = "S3 bucket time, where locking state will be placed"
}