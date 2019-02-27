output "db-name" {
  value = "${aws_rds_cluster.trillian-db-cluster.database_name}"
}
output "db-username" {
  value = "${aws_rds_cluster.trillian-db-cluster.master_username}"
}
output "db-password" {
  value = "${aws_rds_cluster.trillian-db-cluster.master_password}"
}
output "db-endpoint" {
  value = "${aws_rds_cluster.trillian-db-cluster.endpoint}"
}

data "aws_region" "current" {}

output "region" {
  value = "${data.aws_region.current.name}"
}



output "azs" {
  value =  "${data.aws_availability_zones.available.names}"
}
