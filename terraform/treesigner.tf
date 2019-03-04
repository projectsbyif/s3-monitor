resource "aws_lambda_function" "treesigner" {
  filename         = "treesigner.zip"
  function_name    = "treesigner"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "treesigner"
  runtime          = "go1.x"
  source_code_hash = "${base64sha256(file("treesigner.zip"))}"
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

resource "aws_lambda_permission" "cloudwatch_trigger" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.treesigner.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.lambda.arn}"
}

resource "aws_cloudwatch_event_rule" "lambda" {
  name                = "Tree-signer-event"
  description         = "Schedule trigger for lambda execution"
  schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "lambda" {
  target_id = "${aws_lambda_function.treesigner.function_name}"
  rule      = "${aws_cloudwatch_event_rule.lambda.name}"
  arn       = "${aws_lambda_function.treesigner.arn}"
}
