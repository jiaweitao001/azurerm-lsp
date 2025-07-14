package snippets

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	lsp "github.com/Azure/ms-terraform-lsp/internal/protocol"
	provider_schema "github.com/Azure/ms-terraform-lsp/provider-schema"
)

//go:embed msgraph_templates.json
var msgraphTemplateJSON embed.FS

//go:embed azapi_templates.json
var azapiTemplateJSON embed.FS

type CompletionModel struct {
	Label         string             `json:"label"`
	Documentation DocumentationModel `json:"documentation"`
	SortText      string             `json:"sortText"`
	TextEdit      TextEditModel      `json:"textEdit"`
}

type TextEditModel struct {
	NewText string `json:"newText"`
}

type DocumentationModel struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

var (
	msgraphTemplateCandidates []lsp.CompletionItem
	azurermTemplateCandidates []lsp.CompletionItem
	azapiTemplateCandidates   []lsp.CompletionItem
)

func MSGraphTemplateCandidates(editRange lsp.Range) []lsp.CompletionItem {
	if len(msgraphTemplateCandidates) != 0 {
		for i := range msgraphTemplateCandidates {
			msgraphTemplateCandidates[i].TextEdit.Range = editRange
		}
		return msgraphTemplateCandidates
	}
	templates := make([]CompletionModel, 0)
	data, err := msgraphTemplateJSON.ReadFile("msgraph_templates.json")
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return nil
	}

	for _, template := range templates {
		event := lsp.TelemetryEvent{
			Version: lsp.TelemetryFormatVersion,
			Name:    "textDocument/completion",
			Properties: map[string]interface{}{
				"kind":  "template",
				"label": template.Label,
			},
		}
		data, _ := json.Marshal(event)

		newText := strings.ReplaceAll(template.TextEdit.NewText, "$ref", "\\$ref")
		msgraphTemplateCandidates = append(msgraphTemplateCandidates, lsp.CompletionItem{
			Label:  template.Label,
			Kind:   lsp.SnippetCompletion,
			Detail: "Code Sample",
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: template.Documentation.Value,
			},
			SortText:         template.SortText,
			InsertTextFormat: lsp.SnippetTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			TextEdit: &lsp.TextEdit{
				Range:   editRange,
				NewText: newText,
			},
			Command: &lsp.Command{
				Title:     "",
				Command:   "ms-terraform.telemetry",
				Arguments: []json.RawMessage{data},
			},
		})
	}
	return msgraphTemplateCandidates
}

func AzureRMTemplateCandidates(editRange lsp.Range) []lsp.CompletionItem {
	if len(azurermTemplateCandidates) != 0 {
		for i := range azurermTemplateCandidates {
			azurermTemplateCandidates[i].TextEdit.Range = editRange
		}
		return azurermTemplateCandidates
	}

	azurermTemplateCandidates = make([]lsp.CompletionItem, 0)
	for _, obj := range provider_schema.ListAllResourcesAndDataSources() {
		name := obj.Name
		snippet, err := provider_schema.GetSnippet(name, obj.IsDataSource())
		if err != nil {
			continue
		}

		content, err := provider_schema.GetResourceContent(name, obj.IsDataSource())
		if err != nil {
			continue
		}

		kind := "resource"
		if obj.IsDataSource() {
			kind = "data source"
			name = fmt.Sprintf("%s (%s)", obj.GetName(), kind)
		}

		event := lsp.TelemetryEvent{
			Version: lsp.TelemetryFormatVersion,
			Name:    "textDocument/completion",
			Properties: map[string]interface{}{
				"kind": "code-sample",
				"type": name,
			},
		}
		data, _ := json.Marshal(event)

		azurermTemplateCandidates = append(azurermTemplateCandidates, lsp.CompletionItem{
			Label:            name,
			InsertText:       snippet,
			InsertTextFormat: lsp.SnippetTextFormat,
			Kind:             lsp.SnippetCompletion,
			Detail:           "AzureRM " + kind,
			TextEdit: &lsp.TextEdit{
				Range:   editRange,
				NewText: snippet,
			},
			Documentation: lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: content,
			},
			Command: &lsp.Command{
				Title:     "",
				Command:   "ms-terraform.telemetry",
				Arguments: []json.RawMessage{data},
			},
		})
	}
	return azurermTemplateCandidates
}

func AzAPITemplateCandidates(editRange lsp.Range) []lsp.CompletionItem {
	if len(azapiTemplateCandidates) != 0 {
		for i := range azapiTemplateCandidates {
			azapiTemplateCandidates[i].TextEdit.Range = editRange
		}
		return azapiTemplateCandidates
	}
	templates := make([]CompletionModel, 0)
	data, err := azapiTemplateJSON.ReadFile("azapi_templates.json")
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return nil
	}

	for _, template := range templates {
		event := lsp.TelemetryEvent{
			Version: lsp.TelemetryFormatVersion,
			Name:    "textDocument/completion",
			Properties: map[string]interface{}{
				"kind":  "template",
				"label": template.Label,
			},
		}
		data, _ := json.Marshal(event)

		azapiTemplateCandidates = append(azapiTemplateCandidates, lsp.CompletionItem{
			Label:  template.Label,
			Kind:   lsp.SnippetCompletion,
			Detail: "Code Sample",
			Documentation: lsp.MarkupContent{
				Kind:  "markdown",
				Value: template.Documentation.Value,
			},
			SortText:         template.SortText,
			InsertTextFormat: lsp.SnippetTextFormat,
			InsertTextMode:   lsp.AdjustIndentation,
			TextEdit: &lsp.TextEdit{
				Range:   editRange,
				NewText: template.TextEdit.NewText,
			},
			Command: &lsp.Command{
				Title:     "",
				Command:   "azapi.telemetry",
				Arguments: []json.RawMessage{data},
			},
		})
	}
	return azapiTemplateCandidates
}
