resource "aws_secretsmanager_secret" "trigger_workflow" {
  name = var.secretsmanager_secret_name_trigger_workflow
}

resource "aws_secretsmanager_secret_version" "trigger_workflow" {
  secret_id     = aws_secretsmanager_secret.trigger_workflow.id
  secret_string = jsonencode(yamldecode(data.local_file.secret_trigger_workflow.content))
}

data "local_file" "secret_trigger_workflow" {
  filename = "${path.module}/secret_trigger_workflow.yaml"
}
