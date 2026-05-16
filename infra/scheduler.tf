resource "aws_scheduler_schedule" "stop_app" {
  count = var.enable_daily_stop ? 1 : 0

  name                         = "${local.name_prefix}-stop"
  schedule_expression          = var.daily_stop_schedule_expression
  schedule_expression_timezone = var.daily_stop_timezone
  state                        = "ENABLED"
  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:ec2:stopInstances"
    role_arn = aws_iam_role.scheduler_role.arn
    input = jsonencode({
      InstanceIds = [aws_instance.app.id]
    })
  }
}

resource "aws_scheduler_schedule" "start_app" {
  count = var.create_daily_start_schedule ? 1 : 0

  name                         = "${local.name_prefix}-start"
  schedule_expression          = var.daily_start_schedule_expression
  schedule_expression_timezone = var.daily_start_timezone
  state                        = var.enable_daily_start ? "ENABLED" : "DISABLED"
  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:ec2:startInstances"
    role_arn = aws_iam_role.scheduler_role.arn
    input = jsonencode({
      InstanceIds = [aws_instance.app.id]
    })
  }
}
