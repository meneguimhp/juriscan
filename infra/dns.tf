data "aws_route53_zone" "root" {
  count        = var.enable_dns_record ? 1 : 0
  name         = var.root_domain
  private_zone = false
}

resource "aws_route53_record" "app" {
  count   = var.enable_dns_record ? 1 : 0
  zone_id = data.aws_route53_zone.root[0].zone_id
  name    = local.fqdn
  type    = "A"
  ttl     = 300

  records = [var.enable_eip ? aws_eip.app[0].public_ip : aws_instance.app.public_ip]
}
