package handlers

import (
	"context"

	lsctx "github.com/Azure/azurerm-lsp/internal/context"
	"github.com/Azure/azurerm-lsp/internal/langserver/handlers/validate"
	ilsp "github.com/Azure/azurerm-lsp/internal/lsp"
	lsp "github.com/Azure/azurerm-lsp/internal/protocol"
)

func (lh *logHandler) TextDocumentDidOpen(ctx context.Context, params lsp.DidOpenTextDocumentParams) error {
	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return err
	}

	f := ilsp.FileFromDocumentItem(params.TextDocument)
	err = fs.CreateAndOpenDocument(f, f.LanguageID(), f.Text())
	if err != nil {
		return err
	}

	notifier, err := lsctx.DiagnosticsNotifier(ctx)
	if err != nil {
		return err
	}

	diags := validate.NewDiagnostics(f.Text(), f.Filename())
	notifier.PublishHCLDiags(ctx, f.Dir(), diags)
	return nil
}
