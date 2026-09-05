# Security groups placeholder para ALB / tasks ECS / RDS (recursos criados na Fase C).
# Postgres NÃO abre 0.0.0.0/0.

resource "aws_security_group" "alb" {
  name        = "${var.project_name}-alb"
  description = "ALB publico (Fase C). HTTP 80."
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "HTTP da internet (origin CloudFront na Fase C)"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Para targets (tasks) e health checks"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-alb"
  }
}

resource "aws_security_group" "ecs_tasks" {
  name        = "${var.project_name}-ecs-tasks"
  description = "Tasks Fargate (Fase C). Ingress so do ALB na 8080."
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "API a partir do ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    description = "ECR, SSM, RDS, CloudWatch. Sem NAT: task com IP publico (Fase C)."
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-ecs-tasks"
  }
}

resource "aws_security_group" "rds" {
  name        = "${var.project_name}-rds"
  description = "RDS PostgreSQL (Fase C). 5432 somente a partir das tasks."
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "PostgreSQL somente das tasks ECS"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-rds"
  }
}
