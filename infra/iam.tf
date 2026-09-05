# Roles ECS: execução (ECR + awslogs + SSM/KMS) e tarefa (app).

data "aws_iam_policy_document" "ecs_tasks_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_execution" {
  name               = "${var.project_name}-ecs-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume.json
}

resource "aws_iam_role_policy_attachment" "ecs_execution_managed" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "ecs_execution_ssm" {
  statement {
    sid = "SSMParameters"
    actions = [
      "ssm:GetParameters",
      "ssm:GetParameter",
    ]
    resources = [
      aws_ssm_parameter.jwt_secret.arn,
      aws_ssm_parameter.admin_password.arn,
      aws_ssm_parameter.admin_email.arn,
      aws_ssm_parameter.cors_origin.arn,
      aws_ssm_parameter.db_password.arn,
      aws_ssm_parameter.database_url.arn,
    ]
  }

  statement {
    sid       = "KMSDecryptSSM"
    actions   = ["kms:Decrypt"]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["ssm.${var.aws_region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "ecs_execution_ssm" {
  name   = "${var.project_name}-ecs-execution-ssm"
  role   = aws_iam_role.ecs_execution.id
  policy = data.aws_iam_policy_document.ecs_execution_ssm.json
}

resource "aws_iam_role" "ecs_task" {
  name               = "${var.project_name}-ecs-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume.json
}

# Task role: Put/Get/Delete de objetos VOD (sem access key no env).
data "aws_iam_policy_document" "ecs_task_vod_s3" {
  statement {
    sid = "VODObjectRW"
    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObject",
    ]
    resources = ["${aws_s3_bucket.vod.arn}/vod/*"]
  }

  statement {
    sid       = "VODListPrefix"
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.vod.arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["vod/*"]
    }
  }
}

resource "aws_iam_role_policy" "ecs_task_vod_s3" {
  name   = "${var.project_name}-ecs-task-vod-s3"
  role   = aws_iam_role.ecs_task.id
  policy = data.aws_iam_policy_document.ecs_task_vod_s3.json
}
