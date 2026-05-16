output "aws_account_id" {
  description = "Conta AWS atual."
  value       = data.aws_caller_identity.current.account_id
}

output "region" {
  description = "Regiao alvo da stack Juriscan."
  value       = var.region
}

output "vpc_id" {
  value = var.existing_vpc_id
}

output "public_subnet_id" {
  value = var.existing_public_subnet_id
}

output "security_group_id" {
  value = aws_security_group.app.id
}

output "instance_id" {
  value = aws_instance.app.id
}

output "instance_private_ip" {
  value = aws_instance.app.private_ip
}

output "instance_public_ip" {
  value = var.enable_eip ? aws_eip.app[0].public_ip : aws_instance.app.public_ip
}

output "app_fqdn" {
  value = local.fqdn
}

output "dns_record_created" {
  value = var.enable_dns_record
}

output "artifact_bucket_name" {
  value = aws_s3_bucket.artifacts.bucket
}

output "app_deploy_enabled" {
  value = var.deploy_app
}
