package handlers

import (
	"context"
	"fmt"
	"github.com/Azure/azurerm-lsp/provider-schema/azurerm/schema"
	"strings"

	ilsp "github.com/Azure/azurerm-lsp/internal/lsp"
	"github.com/Azure/azurerm-lsp/internal/parser"
	lsp "github.com/Azure/azurerm-lsp/internal/protocol"
	provider_schema "github.com/Azure/azurerm-lsp/provider-schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (svc *service) HandleHover(ctx context.Context, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	docContent, docFileName, err := GetDocumentContent(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %v", err)
	}

	file, diags := hclsyntax.ParseConfig([]byte(docContent), docFileName, hcl.InitialPos)
	if diags != nil && diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse document: %v", diags)
	}

	hclPos := ilsp.LSPPosToHCL(params.Position)
	body, isHcl := file.Body.(*hclsyntax.Body)
	if !isHcl {
		return nil, fmt.Errorf("file is not HCL")
	}

	block := parser.BlockAtPos(body, hclPos)
	if block == nil || len(block.Labels) == 0 || !strings.HasPrefix(block.Labels[0], schema.AzureRMPrefix) {
		return nil, nil
	}

	resourceName := block.Labels[0]
	subBlock := parser.BlockAtPos(block.Body, hclPos)
	attribute := parser.AttributeAtPos(block, hclPos)

	var hoverContent string
	var hoverRange lsp.Range

	if attribute != nil {
		path := buildNestedPath(block, attribute)
		hoverContent, _ = getHoverContent(resourceName, path)
		hoverRange = ilsp.HCLRangeToLSP(attribute.Range())
	} else if subBlock != nil {
		path := buildNestedPath(block, subBlock)
		hoverContent, _ = getHoverContent(resourceName, path)
		hoverRange = ilsp.HCLRangeToLSP(subBlock.Range())
	} else {
		hoverContent, _ = getHoverContentForResource(resourceName)
		hoverRange = ilsp.HCLRangeToLSP(block.Range())
	}

	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: hoverContent,
		},
		Range: hoverRange,
	}, nil
}

func getHoverContentForResource(resourceName string) (string, error) {
	resourceInfo, err := provider_schema.GetObjectInfo(resourceName)
	if err != nil {
		return "", fmt.Errorf("error retrieving resource info: %v", err)
	}

	content := fmt.Sprintf(
		"### %s [Docs](%s) | [Issues](%s)\n\n"+
			"%s\n\n"+
			"**Example HCL**:\n\n"+
			"```hcl\n\n%s\n\n```\n\n"+
			"**Timeouts**: %v\n\n",
		resourceName,
		resourceInfo.GetResourceOrDataSourceDocLink(),
		resourceInfo.GetGitHubIssueLink(),
		resourceInfo.Description,
		resourceInfo.ExampleHCL,
		resourceInfo.Timeouts.String(),
	)
	return content, nil
}

func getHoverContent(resourceName, path string) (string, error) {
	terraformObject, err := provider_schema.GetObjectInfo(resourceName)
	if err != nil {
		return "", fmt.Errorf("error retrieving object info: %v", err)
	}

	propertyInfo, err := provider_schema.GetPropertyInfo(resourceName, path)
	if err != nil {
		return "", fmt.Errorf("error retrieving property info: %v", err)
	}

	// Determine if the property is Required or Optional
	requirementStatus := ""
	if propertyInfo.Required {
		requirementStatus = "Required"
	}
	if propertyInfo.Optional {
		requirementStatus = "Optional"
	}

	detail := propertyInfo.Description
	if propertyInfo.Default != nil {
		detail = fmt.Sprintf("**Default**: %v\n\n", propertyInfo.Default) + detail
	}
	if len(propertyInfo.PossibleValues) > 0 {
		detail += fmt.Sprintf("**Possible Values**: %s\n\n", strings.Join(propertyInfo.PossibleValues, ", "))
	}

	content := fmt.Sprintf(
		"### %s (%s) (%s) [Docs](%s) | [Issues](%s)\n\n %s",
		propertyInfo.Name,
		requirementStatus,
		propertyInfo.AttributeType.FriendlyName(),
		propertyInfo.GetAttributeDocLink(terraformObject.GetResourceOrDataSourceDocLink()),
		propertyInfo.GetGitHubIssueLink(),
		detail,
	)

	return content, nil
}
