# RDS PostgreSQL efêmero — db.t4g.micro, single-AZ, subnet pública, sem ElastiCache (FR-007/008).

resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-public"
  subnet_ids = aws_subnet.public[*].id

  tags = {
    Name = "${var.project_name}-public"
  }
}

resource "aws_db_instance" "main" {
  identifier     = "${var.project_name}-pg"
  engine         = "postgres"
  engine_version = "16"
  instance_class = var.db_instance_class

  db_name  = var.db_name
  username = var.db_username
  password = local.db_password
  port     = 5432

  allocated_storage = 20
  storage_type      = "gp3"
  storage_encrypted     = true

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  multi_az               = false
  availability_zone      = local.azs[0]

  skip_final_snapshot       = true
  deletion_protection       = false
  backup_retention_period   = 0
  delete_automated_backups  = true
  copy_tags_to_snapshot     = false
  apply_immediately         = true
  auto_minor_version_upgrade = true

  performance_insights_enabled = false
  monitoring_interval          = 0

  tags = {
    Name = "${var.project_name}-pg"
  }
}
