package handlers

import (
	"context"

	lsp "github.com/Azure/azurerm-lsp/internal/protocol"
)

func Initialized(ctx context.Context, params lsp.InitializedParams) error {
	return nil
}
