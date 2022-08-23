variable "region" {
  type    = string
  default = "us-east-1"
}

variable "secretsmanager_secret_name" {
  type    = string
  default = "test-gha-dispatcher"
}

variable "zip_path" {
  type        = string
  description = ""
  default     = "gha-dispatcher-lambda_linux_amd64.zip"
}

variable "function_name" {
  type        = string
  description = "Lambda Function Name"
  default     = "test-gha-dispatcher"
}

variable "lambda_role_name" {
  type        = string
  description = ""
  default     = "test-gha-dispatcher"
}
