terraform {
  required_providers {
    azurerm = {
      source = "Azure/azurerm"
    }
  }
}

provider "azurerm" {
  skip_provider_registration = false
}

variable "resource_name" {
  type    = string
  default = "acctest0001"
}

variable "location" {
  type    = string
  default = "westeurope"
}

resource "azurerm_resource" "resourceGroup" {
  type                      = "Microsoft.Resources/resourceGroups@2020-06-01"
  name                      = var.resource_name
  location                  = var.location
}

resource "azurerm_resource" "factory" {
  type      = "Microsoft.DataFactory/factories@2018-06-01"
  parent_id = azurerm_resource.resourceGroup.id
  name      = var.resource_name
  location  = var.location
  body = jsonencode({
    properties = {
      publicNetworkAccess = "Enabled"
      repoConfiguration   = null
    }
  })
  schema_validation_enabled = false
  response_export_values    = ["*"]
}

resource "azurerm_resource" "integrationRuntime" {
  type      = "Microsoft.DataFactory/factories/integrationRuntimes@2018-06-01"
  parent_id = azurerm_resource.factory.id
  name      = var.resource_name
  body = jsonencode({
    properties = {
      description = ""
      type        = "SelfHosted"
    }
  })
  schema_validation_enabled = false
  response_export_values    = ["*"]
}

