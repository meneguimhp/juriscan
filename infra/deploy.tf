resource "terraform_data" "deploy_app" {
  count = var.deploy_app ? 1 : 0

  triggers_replace = {
    backend_hash = sha256(join("", [
      for source_file in fileset("${path.module}/../juriscan-backend", "**/*.go") :
      filesha256("${path.module}/../juriscan-backend/${source_file}")
    ]))
    frontend_hash = sha256(join("", [
      for source_file in fileset("${path.module}/../juriscan-frontend/src", "**/*") :
      filesha256("${path.module}/../juriscan-frontend/src/${source_file}")
    ]))
    frontend_public_hash = sha256(join("", [
      for source_file in fileset("${path.module}/../juriscan-frontend/public", "**/*") :
      filesha256("${path.module}/../juriscan-frontend/public/${source_file}")
    ]))
    package_hash = filesha256("${path.module}/../juriscan-frontend/package-lock.json")
    deploy_hash  = filesha256("${path.module}/scripts/deploy-app.ps1")
    instance_id  = aws_instance.app.id
    admin_emails = join(",", var.bootstrap_admin_emails)
    token_echo   = tostring(var.login_token_echo)
    reset_mysql  = tostring(var.reset_mysql_volume_on_deploy)
  }

  provisioner "local-exec" {
    interpreter = ["PowerShell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File"]
    command     = "${path.module}/scripts/deploy-app.ps1"

    environment = {
      AWS_REGION          = var.region
      INSTANCE_ID         = aws_instance.app.id
      ARTIFACT_BUCKET     = aws_s3_bucket.artifacts.bucket
      APP_FQDN            = local.fqdn
      MYSQL_APP_PASSWORD  = random_password.mysql_app_password.result
      MYSQL_ROOT_PASSWORD = random_password.mysql_root_password.result
      ADMIN_EMAILS        = join(",", var.bootstrap_admin_emails)
      LOGIN_TOKEN_ECHO    = tostring(var.login_token_echo)
      WHATSAPP_PROVIDER   = "mock"
      RESET_MYSQL_VOLUME  = tostring(var.reset_mysql_volume_on_deploy)
    }
  }

  depends_on = [
    aws_eip_association.app,
    aws_iam_role_policy.ec2_artifact_read,
    aws_s3_bucket_public_access_block.artifacts,
    aws_s3_bucket_server_side_encryption_configuration.artifacts,
    aws_s3_bucket_versioning.artifacts
  ]
}
