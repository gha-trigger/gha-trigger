variable "region" {
  type    = string
  default = "us-east-1"
}

variable "secretsmanager_secret_name_main" {
  type    = string
  default = "test-gha-trigger-main"
}

variable "secretsmanager_secret_name_trigger_workflow" {
  type    = string
  default = "test-gha-trigger-trigger-workflow"
}

variable "lambda_architecture" {
  type        = string
  description = ""
  default     = "arm64"
}

variable "zip_path" {
  type        = string
  description = ""
  default     = "gha-trigger-lambda_linux_arm64.zip"
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

variable "api_gateway_name" {
  type        = string
  description = ""
  default     = "test-gha-trigger"
}
