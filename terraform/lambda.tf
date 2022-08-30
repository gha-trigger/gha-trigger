resource "aws_lambda_function" "main" {
  filename         = var.zip_path
  function_name    = var.function_name
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  source_code_hash = filebase64sha256(var.zip_path)
  runtime          = "provided.al2"
  timeout          = 300
  architectures    = [var.lambda_architecture]

  environment {
    variables = {
      CONFIG = file("${path.module}/config.yaml")
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = var.lambda_role_name
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.lambda.json
}

data "aws_iam_policy_document" "lambda" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_cloudwatch_log_group" "lambda_log" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = 7
}

resource "aws_iam_role_policy" "lambda_log" {
  name   = "log"
  policy = data.aws_iam_policy_document.lambda_log.json
  role   = aws_iam_role.lambda.name
}

data "aws_iam_policy_document" "lambda_log" {
  statement {
    actions   = ["logs:CreateLogStream"]
    resources = ["${aws_cloudwatch_log_group.lambda_log.arn}:log-stream:*"]
  }
  statement {
    actions   = ["logs:PutLogEvents"]
    resources = ["*"]
  }
}
