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

resource "aws_security_group" "trillian-s3-handler" {
  name        = "Trillian S3 event handler"
  vpc_id      = "${aws_vpc.main.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
