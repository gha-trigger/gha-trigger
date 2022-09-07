resource "aws_api_gateway_rest_api" "main" {
  name = "test-gha-trigger"
}

resource "aws_api_gateway_stage" "default" {
  deployment_id = aws_api_gateway_deployment.main.id
  rest_api_id   = aws_api_gateway_rest_api.main.id
  stage_name    = "main"
}

resource "aws_api_gateway_resource" "main" {
  path_part   = "webhook"
  parent_id   = aws_api_gateway_rest_api.main.root_resource_id
  rest_api_id = aws_api_gateway_rest_api.main.id
}

resource "aws_api_gateway_method" "main" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.main.id
  http_method   = "POST"
  authorization = "NONE"
  request_parameters = {
    "method.request.header.X-GitHub-Hook-Installation-Target-ID" = true
  }
}

resource "aws_api_gateway_integration" "main" {
  rest_api_id             = aws_api_gateway_rest_api.main.id
  resource_id             = aws_api_gateway_resource.main.id
  http_method             = aws_api_gateway_method.main.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = aws_lambda_function.main.invoke_arn

  request_parameters = {
    # Invoke Lambda Function asynchronously
    # https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-integration-async.html
    # single quotes 'Event' are required
    "integration.request.header.X-Amz-Invocation-Type" = "'Event'"
  }

  request_templates = {
    "application/json" = <<-EOT
      ##  See http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html
      ##  This template will pass through all parameters including path, querystring, header, stage variables, and context through to the integration endpoint via the body/payload
      #set($allParams = $input.params())
      {
      "body-json" : "$util.escapeJavaScript($input.body)",
      "params" : {
      #foreach($type in $allParams.keySet())
          #set($params = $allParams.get($type))
      "$type" : {
          #foreach($paramName in $params.keySet())
          "$paramName" : "$util.escapeJavaScript($params.get($paramName))"
              #if($foreach.hasNext),#end
          #end
      }
          #if($foreach.hasNext),#end
      #end
      },
      "stage-variables" : {
      #foreach($key in $stageVariables.keySet())
      "$key" : "$util.escapeJavaScript($stageVariables.get($key))"
          #if($foreach.hasNext),#end
      #end
      },
      "context" : {
          "account-id" : "$context.identity.accountId",
          "api-id" : "$context.apiId",
          "api-key" : "$context.identity.apiKey",
          "authorizer-principal-id" : "$context.authorizer.principalId",
          "caller" : "$context.identity.caller",
          "cognito-authentication-provider" : "$context.identity.cognitoAuthenticationProvider",
          "cognito-authentication-type" : "$context.identity.cognitoAuthenticationType",
          "cognito-identity-id" : "$context.identity.cognitoIdentityId",
          "cognito-identity-pool-id" : "$context.identity.cognitoIdentityPoolId",
          "http-method" : "$context.httpMethod",
          "stage" : "$context.stage",
          "source-ip" : "$context.identity.sourceIp",
          "user" : "$context.identity.user",
          "user-agent" : "$context.identity.userAgent",
          "user-arn" : "$context.identity.userArn",
          "request-id" : "$context.requestId",
          "resource-id" : "$context.resourceId",
          "resource-path" : "$context.resourcePath"
          }
      }
    EOT
  }
}

resource "aws_api_gateway_method_response" "main" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  resource_id = aws_api_gateway_resource.main.id
  http_method = aws_api_gateway_method.main.http_method
  status_code = "202"
}

resource "aws_api_gateway_integration_response" "main" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  resource_id = aws_api_gateway_resource.main.id
  http_method = aws_api_gateway_method.main.http_method
  status_code = aws_api_gateway_method_response.main.status_code
  depends_on = [
    # https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_integration_response
    # > Depends on having aws_api_gateway_integration inside your rest api.
    # To ensure this you might need to add an explicit depends_on for clean runs.
    aws_api_gateway_integration.main
  ]
}

resource "aws_api_gateway_deployment" "main" {
  rest_api_id       = aws_api_gateway_rest_api.main.id
  stage_description = "setting file hash = ${md5(file("api_gateway.tf"))}"

  depends_on = [
    # Error: Error creating API Gateway Deployment: BadRequestException: The REST API doesn't contain any methods
    aws_api_gateway_method.main,
    # Error: Error creating API Gateway Deployment: BadRequestException: No integration defined for method
    aws_api_gateway_integration.main
  ]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lambda_permission" "main" {
  statement_id  = "AllowLambuildInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"

  # More: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html
  source_arn = "arn:aws:execute-api:${var.region}:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.main.id}/*/${aws_api_gateway_method.main.http_method}${aws_api_gateway_resource.main.path}"
}
