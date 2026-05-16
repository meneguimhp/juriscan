resource "aws_instance" "app" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = var.existing_public_subnet_id
  vpc_security_group_ids = [aws_security_group.app.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2_ssm_profile.name
  key_name               = var.key_name != "" ? var.key_name : null

  user_data = <<-EOT
    #!/bin/bash
    set -euxo pipefail

    dnf update -y
    dnf install -y docker
    systemctl enable docker
    systemctl start docker

    mkdir -p /opt/juriscan
    chown ec2-user:ec2-user /opt/juriscan

    # plugin docker compose (amazon linux 2023)
    dnf install -y docker-compose-plugin || true
  EOT

  root_block_device {
    volume_size           = var.root_volume_size
    volume_type           = "gp3"
    delete_on_termination = true
  }

  lifecycle {
    ignore_changes = [ami]
  }

  tags = {
    Name = "${local.name_prefix}-app"
  }
}

resource "aws_eip" "app" {
  count  = var.enable_eip ? 1 : 0
  domain = "vpc"

  tags = {
    Name = "${local.name_prefix}-eip"
  }
}

resource "aws_eip_association" "app" {
  count         = var.enable_eip ? 1 : 0
  instance_id   = aws_instance.app.id
  allocation_id = aws_eip.app[0].id
}
