variable "location" {}

variable "subscription_id" {
  type        = string
  description = "Azure Subscription ID"
}

variable "function_app" {
  type        = string
  description = "Azure Function App Name"
}

variable "tags" {
  type = map(any)

  default = {
    Environment = "Dev"
    Stack       = "Go"
  }
}

# SharePoint Bindings

variable "sp_source_creds" {
  type        = string
  description = "Source SharePoint Creds"
}

variable "sp_target_creds" {
  type        = string
  description = "Target SharePoint Cresa"
}

variable "sp_master_key" {
  type        = string
  description = "Secret Encryption Key"
}

# Custom handlers package

variable "package" {
  type    = string
  default = "./package/functions.zip"
}
