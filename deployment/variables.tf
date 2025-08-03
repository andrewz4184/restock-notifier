variable "lambda_role_arn" {
  type = string
}

variable "telegram_bot_token" {
  type        = string
  description = "Telegram Bot API Token"
}

variable "telegram_chat_id" {
  type        = string
  description = "Telegram Chat ID to send messages to"
}
