# Cluster + task Fargate ARM 256/512 + service desired_count=1, IP público, sem NAT (research §3).
# Imagem ECR :latest vazia até a Fase D — health /health pode falhar.

resource "aws_ecs_cluster" "main" {
  name = var.project_name

  setting {
    name  = "containerInsights"
    value = "disabled"
  }

  tags = {
    Name = var.project_name
  }
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${var.project_name}-api"
  retention_in_days = 7
}

resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project_name}-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.fargate_cpu
  memory                   = var.fargate_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  container_definitions = jsonencode([{
    name      = "api"
    image     = "${aws_ecr_repository.api.repository_url}:latest"
    essential = true
    portMappings = [{
      containerPort = 8080
      hostPort      = 8080
      protocol      = "tcp"
    }]
    environment = [
      { name = "PORT", value = "8080" },
      { name = "JWT_EXPIRATION", value = "24h" },
      { name = "VOD_BACKEND", value = "s3" },
      { name = "VOD_S3_BUCKET", value = aws_s3_bucket.vod.id },
      { name = "VOD_PLAYBACK_TTL", value = "15m" },
      { name = "AWS_REGION", value = var.aws_region },
    ]
    secrets = [
      { name = "DATABASE_URL", valueFrom = aws_ssm_parameter.database_url.arn },
      { name = "JWT_SECRET", valueFrom = aws_ssm_parameter.jwt_secret.arn },
      { name = "ADMIN_PASSWORD", valueFrom = aws_ssm_parameter.admin_password.arn },
      { name = "ADMIN_EMAIL", valueFrom = aws_ssm_parameter.admin_email.arn },
      { name = "CORS_ORIGIN", valueFrom = aws_ssm_parameter.cors_origin.arn },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.api.name
        awslogs-region        = var.aws_region
        awslogs-stream-prefix = "api"
      }
    }
  }])
}

resource "aws_ecs_service" "api" {
  name            = "${var.project_name}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.ecs_desired_count
  launch_type     = "FARGATE"
  propagate_tags  = "SERVICE"

  deployment_minimum_healthy_percent = 0
  deployment_maximum_percent         = 200
  health_check_grace_period_seconds  = 60
  wait_for_steady_state              = false

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener.http]
}
