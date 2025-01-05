resource "azurerm_resource" "storage" {
  type = "Microsoft.App/managedEnvironments/storages@2022-03-01"
  parent_id = azurerm_resource.managedEnvironment.id
  name = "henglu9886"
  body = jsonencode({
    properties = {
      azureFile = {
        accessMode = "ReadWrite"
        accountKey = jsondecode(data.azurerm_resource_action.listKeys.output).keys[0].value
        accountName = "henglu1360"
        shareName = "testsharehkez7"
      }
    }
  })
}