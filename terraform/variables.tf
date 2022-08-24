variable "region" {
  type    = string
  default = "us-east-1"
}

variable "secretsmanager_secret_name" {
  type    = string
  default = "test-gha-trigger"
}

variable "zip_path" {
  type        = string
  description = ""
  default     = "gha-trigger-lambda_linux_amd64.zip"
}

variable "function_name" {
  type        = string
  description = "Lambda Function Name"
  default     = "test-gha-trigger"
}

variable "lambda_role_name" {
  type        = string
  description = ""
  default     = "test-gha-trigger"
}
