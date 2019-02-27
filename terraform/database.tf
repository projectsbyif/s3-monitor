resource "aws_db_subnet_group" "trillian-db" {
  name       = "trillian-db"
  subnet_ids = ["${aws_subnet.main-a.id}", "${aws_subnet.main-b.id}", "${aws_subnet.main-c.id}"]
}

resource "aws_rds_cluster" "trillian-db-cluster" {
  cluster_identifier      = "aurora-cluster"
  engine                  = "aurora"
  availability_zones      = ["${data.aws_availability_zones.available.names}"]
  database_name           = "trillian"
  master_username         = "trillian"
  master_password         = "trillian"
  engine_mode             = "serverless"
  backup_retention_period = 1
  db_subnet_group_name    = "${aws_db_subnet_group.trillian-db.name}"
  vpc_security_group_ids  = ["${aws_security_group.trillian-db.id}"]

  scaling_configuration {
    auto_pause               = true
    max_capacity             = 2
    min_capacity             = 2
    seconds_until_auto_pause = 300
  }
}

resource "aws_security_group" "trillian-db" {
  vpc_id      = "${aws_vpc.main.id}"
  name        = "trillian-db"

  ingress {
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = ["${aws_security_group.trillian-s3-handler.id}"]
  }
}
