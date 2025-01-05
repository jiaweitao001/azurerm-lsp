# Terraform azurerm Provider Language Server

Experimental version of [terraform-provider-azurerm](https://github.com/hashicorp/terraform-provider-azurerm) language server.

## What is LSP

Read more about the Language Server Protocol at https://microsoft.github.io/language-server-protocol/

## Introduction

This project only supports completion/hover/diagnostics for `terraform-provider-azurerm`,
not targeting support all language features for `HCL` or `Terraform`. To get the best user experience, 
it's recommended to use it with language server for `Terraform`.

## Features

- Completion of `azurerm` resources
- Completion of allowed azure resource types when input `type` in `azurerm` resources
- Completion of allowed azure resource properties when input `body` in `azurerm` resources, limitation: it only works when use `jsonencode` function to build the JSON
- Better completion for discriminated object
- Completion for all required properties
- Show hint when hover on `azurerm` resources
- Show diagnostics for properties defined inside `body`

## Installation

1. Clone this project to local
2. Run `go install` under the project folder.

## Usage

The most reasonable way you will interact with the language server
is through a client represented by an IDE, or a plugin of an IDE.

VSCode extension: [vscode-azureterraform](https://github.com/Azure/vscode-azureterraform)

## Credits

We wish to thank HashiCorp for the use of some MPLv2-licensed code from their open source project [terraform-ls](https://github.com/hashicorp/terraform-ls).

