provider "aws" {
  region  = "ap-southeast-1"
  profile = "default"
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
    }
  }
}

resource "aws_cloudwatch_event_rule" "matcha_farmer_schedule" {
  name                = "matcha_farmer_schedule"
  description         = "Trigger at 1, 2, 3 PM JST daily"
  schedule_expression = "cron(0 4,5,6 * * ? *)"
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

