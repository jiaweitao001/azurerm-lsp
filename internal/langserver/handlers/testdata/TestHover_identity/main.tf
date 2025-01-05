resource "azurerm_resource" "cluster" {
  type      = "Microsoft.ContainerService/managedClusters@2024-02-01"
  parent_id = azurerm_resource.resourceGroup.id
  name      = "example"
  location  = azurerm_resource.resourceGroup.location
  identity {
    type         = "SystemAssigned"
    identity_ids = []
  }
  body = {
    properties = {
      agentPoolProfiles = [
        {
          count  = 1
          mode   = "System"
          name   = "default"
          vmSize = "Standard_DS2_v2"
        },
      ]
      dnsPrefix = "example"
    }
  }
}