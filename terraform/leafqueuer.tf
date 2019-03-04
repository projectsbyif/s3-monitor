resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.leafqueuer.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${aws_s3_bucket.logs.arn}"
}

resource "aws_lambda_function" "leafqueuer" {
  filename         = "leafqueuer.zip"
  function_name    = "leafqueuer"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "leafqueuer"
  runtime          = "go1.x"
  source_code_hash = "${base64sha256(file("leafqueuer.zip"))}"
  timeout          = 60 // Increased timeout because waking up the serverless Aurora can take some time

  environment      = {
    variables = {
      LAMBDA_LOGTOSTDERR          = "true"
      LAMBDA_TREEID               = "${var.tree_id}"
      LAMBDA_MYSQL_URI            = "${aws_rds_cluster.trillian-db-cluster.master_username}:${aws_rds_cluster.trillian-db-cluster.master_password}@tcp(${aws_rds_cluster.trillian-db-cluster.endpoint}:${aws_rds_cluster.trillian-db-cluster.port})/${aws_rds_cluster.trillian-db-cluster.database_name}"
      LAMBDA_MYSQL_MAX_IDLE_CONNS = 0
    }
  }

  vpc_config    = {
    subnet_ids         = ["${aws_subnet.main-a.id}", "${aws_subnet.main-b.id}", "${aws_subnet.main-c.id}"]
    security_group_ids = ["${aws_security_group.trillian-s3-handler.id}"]
  }
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = "${aws_s3_bucket.logs.id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.leafqueuer.arn}"
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
  }
}
