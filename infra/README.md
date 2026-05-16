# Juriscan Infra (AWS single instance / baixo custo)

Este diretorio contem todo o Terraform do Juriscan para subir uma EC2 unica de baixo custo.
Ele reaproveita uma VPC/subnet publica existente, por exemplo a do SynPlace, mas cria recursos proprios do Juriscan.

## O que cria
- Security Group proprio do Juriscan (80/443 publicos; SSH opcional).
- 1 EC2 Amazon Linux 2023 com Docker instalado via user_data.
- IAM + SSM para operacao sem SSH.
- Elastic IP (opcional, padrao ligado).
- Schedules de stop/start (start criado desabilitado por padrao).
- Budget mensal opcional.
- DNS A record opcional (`juriscan.synplace.com.br`).

## Isolamento
- Regiao padrao: `us-east-1`.
- Prefixo de recursos: `juriscan-prod-*`.
- State dedicado em bucket separado do SynPlace.
- Nao altera VPC, subnets, instancias, Security Groups ou registros existentes do SynPlace.
- Cria apenas recursos novos do Juriscan e, se habilitado, o registro `juriscan.synplace.com.br`.

## Pre-requisito
Criar bucket de state antes do `init`:
- bucket: `terraform-juriscan-state`
- region: `us-east-1`
- versioning + encryption + block public access

## Uso
```bash
cd infra
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform plan
terraform apply
```

Preencha no `terraform.tfvars`:
- `existing_vpc_id`
- `existing_public_subnet_id`

## Custos (perfil baixo)
- Instancia padrao: `t3.micro`
- Stop diario as 22:00 (enabled)
- Start diario as 08:00 (schedule criado, disabled)

## Publicacao de dominio
Quando quiser publicar:
- `enable_dns_record = true`
- `app_subdomain = "juriscan"`

Isso cria `juriscan.synplace.com.br` apontando para a EC2 (EIP se habilitado).

## GitHub
Versionar este diretorio inteiro, exceto arquivos locais ignorados:
- `.terraform/`
- `terraform.tfvars`
- arquivos `*.tfstate*`
