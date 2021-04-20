variable "app_id" {}
variable "api_key" {}

provider "algolia" {
  app_id  = var.app_id
  api_key = var.api_key
}