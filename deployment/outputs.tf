output "lambda_function_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.matcha_farmer.arn
}

output "cloudwatch_event_rule_arn" {
  description = "The ARN of the CloudWatch Event Rule"
  value       = aws_cloudwatch_event_rule.matcha_farmer_schedule.arn
}

output "cloudwatch_event_schedule" {
  description = "The schedule expression for the CloudWatch Event Rule"
  value       = aws_cloudwatch_event_rule.matcha_farmer_schedule.schedule_expression
}
