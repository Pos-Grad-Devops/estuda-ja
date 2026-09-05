# Canal IVS efêmero da sessão (contracts/terraform-ivs.md / research §1).
# BASIC + LOW, sem authorized, sem recording. Destroy remove o canal.
# Stream key padrão do CreateChannel via data source (sem segunda key).

resource "aws_ivs_channel" "live" {
  name         = "${var.project_name}-live"
  type         = "BASIC"
  latency_mode = "LOW"
  authorized   = false

  tags = {
    Name = "${var.project_name}-live"
  }
}

# Key criada junto com o canal — não provisionar aws_ivs_stream_key extra.
data "aws_ivs_stream_key" "live" {
  channel_arn = aws_ivs_channel.live.arn
}
