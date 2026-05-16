variable "region" {
  description = "Regiao AWS onde esta a VPC/subnet publica reaproveitada."
  type        = string
  default     = "us-east-1"
}

variable "project" {
  description = "Nome curto do projeto usado em tags e nomes de recursos."
  type        = string
  default     = "juriscan"
}

variable "environment" {
  description = "Ambiente do Juriscan."
  type        = string
  default     = "prod"
}

variable "root_domain" {
  description = "Dominio raiz hospedado no Route53."
  type        = string
  default     = "synplace.com.br"
}

variable "app_subdomain" {
  description = "Subdominio da aplicacao Juriscan."
  type        = string
  default     = "juriscan"
}

variable "enable_dns_record" {
  description = "Se true, cria/atualiza registro DNS publico do Juriscan."
  type        = bool
  default     = false
}

variable "existing_vpc_id" {
  description = "ID da VPC publica existente a ser reaproveitada (ex.: VPC do SynPlace)."
  type        = string
  nullable    = false

  validation {
    condition     = length(trimspace(var.existing_vpc_id)) > 0
    error_message = "Informe existing_vpc_id no terraform.tfvars."
  }
}

variable "existing_public_subnet_id" {
  description = "ID da subnet publica existente onde a EC2 exclusiva do Juriscan sera criada."
  type        = string
  nullable    = false

  validation {
    condition     = length(trimspace(var.existing_public_subnet_id)) > 0
    error_message = "Informe existing_public_subnet_id no terraform.tfvars."
  }
}

variable "instance_type" {
  description = "Tipo da EC2 unica."
  type        = string
  default     = "t3.micro"
}

variable "root_volume_size" {
  description = "Tamanho do volume raiz da EC2 em GB."
  type        = number
  default     = 20
}

variable "key_name" {
  description = "Key Pair EC2 opcional. Vazio = sem SSH por chave."
  type        = string
  default     = ""
}

variable "ssh_allowed_cidrs" {
  description = "CIDRs autorizados para SSH. Vazio mantem SSH fechado e operacao via SSM."
  type        = list(string)
  default     = []
}

variable "enable_eip" {
  description = "Se true, aloca e associa Elastic IP para IP publico estavel."
  type        = bool
  default     = true
}

variable "enable_daily_stop" {
  description = "Habilita desligamento diario da EC2."
  type        = bool
  default     = true
}

variable "daily_stop_schedule_expression" {
  description = "Expressao cron do EventBridge Scheduler para desligar a EC2."
  type        = string
  default     = "cron(0 22 * * ? *)"
}

variable "daily_stop_timezone" {
  description = "Timezone do desligamento diario."
  type        = string
  default     = "America/Sao_Paulo"
}

variable "create_daily_start_schedule" {
  description = "Se true, cria o schedule de start."
  type        = bool
  default     = true
}

variable "enable_daily_start" {
  description = "Se true, deixa o schedule de start habilitado."
  type        = bool
  default     = false
}

variable "daily_start_schedule_expression" {
  description = "Expressao cron do EventBridge Scheduler para ligar a EC2."
  type        = string
  default     = "cron(0 8 * * ? *)"
}

variable "daily_start_timezone" {
  description = "Timezone do ligamento diario."
  type        = string
  default     = "America/Sao_Paulo"
}

variable "enable_budget" {
  description = "Se true, cria um budget mensal para controle de custo."
  type        = bool
  default     = false
}

variable "monthly_budget_limit_usd" {
  description = "Limite mensal em USD."
  type        = number
  default     = 15
}

variable "budget_alert_emails" {
  description = "Emails para alerta de budget."
  type        = list(string)
  default     = []
}

variable "deploy_app" {
  description = "Se true, Terraform dispara build local, upload para S3 e deploy via SSM na EC2."
  type        = bool
  default     = true
}

variable "artifact_bucket_force_destroy" {
  description = "Permite destruir bucket de artefatos mesmo com objetos."
  type        = bool
  default     = true
}

variable "bootstrap_admin_emails" {
  description = "Emails admin criados no bootstrap do backend."
  type        = list(string)
  default     = ["admin@juriscan.local"]
}

variable "login_token_echo" {
  description = "Mostra token OTP na resposta de login. Usar apenas para piloto sem provedor de email."
  type        = bool
  default     = true
}

variable "reset_mysql_volume_on_deploy" {
  description = "Remove o volume local do MySQL durante deploy. Usar apenas no piloto inicial ou quando dados forem descartaveis."
  type        = bool
  default     = false
}
