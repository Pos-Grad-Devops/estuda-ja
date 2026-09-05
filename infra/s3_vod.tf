# Bucket S3 efêmero para objetos VOD (demo). Distinto do frontend.
# Sem website, sem OAC, sem CloudFront de mídia (research §3 / contracts/terraform-vod.md).

resource "aws_s3_bucket" "vod" {
  bucket        = "${var.project_name}-vod-${data.aws_caller_identity.current.account_id}"
  force_destroy = true

  tags = {
    Name = "${var.project_name}-vod"
  }
}

resource "aws_s3_bucket_ownership_controls" "vod" {
  bucket = aws_s3_bucket.vod.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_public_access_block" "vod" {
  bucket                  = aws_s3_bucket.vod.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "vod" {
  bucket = aws_s3_bucket.vod.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# CORS mínimo para <video> no browser (origem = CloudFront do frontend).
resource "aws_s3_bucket_cors_configuration" "vod" {
  bucket = aws_s3_bucket.vod.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["https://${aws_cloudfront_distribution.frontend.domain_name}"]
    expose_headers  = ["ETag", "Content-Length", "Content-Type", "Accept-Ranges"]
    max_age_seconds = 3000
  }
}
