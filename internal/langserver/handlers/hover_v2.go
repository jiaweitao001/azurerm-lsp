package handlers

import (
	"context"

	"github.com/Azure/azurerm-lsp/internal/parser"
	"github.com/Azure/azurerm-lsp/internal/protocol"
	"github.com/Azure/azurerm-lsp/provider-schema"
)

func (svc *service) HandleHover(ctx context.Context, params protocol.TextDocumentPositionParams) (*protocol.Hover, error) {
	docContent, docFileName, err := parser.GetDocumentContent(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	ctxInfo, diags, err := parser.BuildHCLContext(docContent, docFileName, params.Position)
	if err != nil || (diags != nil && diags.HasErrors()) {
		docContent, fieldName, _, err := parser.AttemptReparse(docContent, params.Position.Line)
		if err != nil {
			return nil, nil
		}

		ctxInfo, diags, err = parser.BuildHCLContext(docContent, docFileName, params.Position)
		if err != nil || (diags != nil && diags.HasErrors()) {
			return nil, nil
		}

		ctxInfo.ParsedPath += "." + fieldName
	}

	var content string

	switch {
	case ctxInfo.ParsedPath != "" && (ctxInfo.SubBlock != nil || ctxInfo.Attribute != nil):
		content, err = provider_schema.GetAttributeContent(ctxInfo.Resource, ctxInfo.ParsedPath)
	default:
		content, _, err = provider_schema.GetResourceContent(ctxInfo.Resource)
	}
	if err != nil {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: content,
		},
	}, nil
}
