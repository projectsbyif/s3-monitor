resource "aws_iam_role" "iam_for_lambda" {
  name        = "iam_for_lambda"
  # description = ""

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "trillian_handler" {
  name   = "trillian_lambda_policy"
  role   = "${aws_iam_role.iam_for_lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
        "Action": [
            "logs:*",
            "ec2:*"
        ],
        "Effect": "Allow",
        "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.func.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${aws_s3_bucket.logs.arn}"
}

resource "aws_lambda_function" "func" {
  filename      = "handler.zip"
  function_name = "handler"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "handler"
  runtime       = "go1.x"

  vpc_config    = {
    subnet_ids         = ["${aws_subnet.main-a.id}", "${aws_subnet.main-b.id}", "${aws_subnet.main-c.id}"]
    security_group_ids = ["${aws_security_group.trillian-s3-handler.id}"]
  }
  //source_code_hash = "${base64sha256(file("handler.zip"))}" TODO - add this once the network / db is configured
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = "${aws_s3_bucket.logs.id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.func.arn}"
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
  }
}

resource "aws_security_group" "trillian-s3-handler" {
  name        = "Trillian S3 event handler"
  vpc_id      = "${aws_vpc.main.id}"
  egress {
    # Connect to the trillian db
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = ["${aws_security_group.trillian-db.id}"]
  }
}
