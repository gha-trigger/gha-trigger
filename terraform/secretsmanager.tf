resource "aws_secretsmanager_secret" "main" {
  name = var.secretsmanager_secret_name_main
}

resource "aws_secretsmanager_secret_version" "main" {
  secret_id     = aws_secretsmanager_secret.main.id
  secret_string = jsonencode(yamldecode(data.local_file.secret.content))
}

data "local_file" "secret" {
  filename = "${path.module}/secret.yaml"
}

resource "aws_iam_role_policy" "read_secret" {
  name   = "read-secret"
  policy = data.aws_iam_policy_document.read_secret.json
  role   = aws_iam_role.lambda.name
}

data "aws_iam_policy_document" "read_secret" {
  statement {
    actions = ["secretsmanager:GetSecretValue"]
    resources = [
      aws_secretsmanager_secret_version.main.arn,
      aws_secretsmanager_secret_version.trigger_workflow.arn,
    ]
  }
}
