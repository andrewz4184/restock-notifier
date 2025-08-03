provider "aws" {
  region  = "ap-southeast-1"
  profile = "default"
}


# DynamoDB table for tracking notification state
resource "aws_dynamodb_table" "matcha_notification_state" {
  name           = "matcha-notification-state"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "date"

  attribute {
    name = "date"
    type = "S"
  }

  tags = {
    Name = "Matcha Notification State"
  }
}

resource "aws_lambda_function" "matcha_farmer" {
  function_name    = "matcha_farmer"
  filename         = "../lambda.zip"
  source_code_hash = filebase64sha256("../lambda.zip")
  handler          = "main"
  runtime          = "provided.al2"
  role             = var.lambda_role_arn
  timeout          = 30
  memory_size      = 128

  environment {
    variables = {
      TELEGRAM_BOT_TOKEN = var.telegram_bot_token
      TELEGRAM_CHAT_ID   = var.telegram_chat_id
      DYNAMODB_TABLE     = aws_dynamodb_table.matcha_notification_state.name
    }
  }
}

resource "aws_cloudwatch_event_rule" "matcha_farmer_schedule" {
  name                = "matcha_farmer_schedule"
  description         = "Trigger every minute during 1-3 PM JST (4-6 AM UTC)"
  schedule_expression = "cron(* 4-6 * * ? *)"
}

resource "aws_cloudwatch_event_target" "matcha_farmer_target" {
  rule      = aws_cloudwatch_event_rule.matcha_farmer_schedule.name
  target_id = "matcha_farmer"
  arn       = aws_lambda_function.matcha_farmer.arn
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.matcha_farmer.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.matcha_farmer_schedule.arn
}

