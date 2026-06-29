package notify

// . ݁₊ ✶. ݁ ˖ email template API — stations customize their email wording

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"radioooooo/internal/auth"
)

type TemplateHandler struct {
	store *TemplateStore
}

func NewTemplateHandler(store *TemplateStore) *TemplateHandler {
	return &TemplateHandler{store: store}
}

func (h *TemplateHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-email-template",
		Method:      http.MethodGet,
		Path:        "/stations/{id}/email-templates/{type}",
		Summary:     "Get email template (custom or default)",
		Tags:        []string{"Email Templates"},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "upsert-email-template",
		Method:      http.MethodPut,
		Path:        "/stations/{id}/email-templates/{type}",
		Summary:     "Set custom email template for a station",
		Tags:        []string{"Email Templates"},
	}, h.upsert)
}

type templateGetInput struct {
	ID   string `path:"id"`
	Type string `path:"type"`
}

type templateGetOutput struct {
	Body EmailTemplate
}

type templateUpsertBody struct {
	Subject string `json:"subject" minLength:"1"`
	Body    string `json:"body"    minLength:"1"`
}

type templateUpsertInput struct {
	ID   string `path:"id"`
	Type string `path:"type"`
	Body templateUpsertBody
}

type templateUpsertOutput struct {
	Body EmailTemplate
}

func (h *TemplateHandler) get(ctx context.Context, input *templateGetInput) (*templateGetOutput, error) {
	if _, ok := auth.StationIDFromContext(ctx); !ok {
		return nil, huma.Error403Forbidden("not authenticated")
	}
	tmpl, err := h.store.Get(ctx, input.ID, input.Type)
	if err != nil {
		return nil, huma.Error404NotFound("unknown template type")
	}
	return &templateGetOutput{Body: tmpl}, nil
}

func (h *TemplateHandler) upsert(ctx context.Context, input *templateUpsertInput) (*templateUpsertOutput, error) {
	stationID, ok := auth.StationIDFromContext(ctx)
	if !ok || stationID != input.ID {
		return nil, huma.Error403Forbidden("not authenticated")
	}
	if _, ok := DefaultTemplate(input.Type); !ok {
		return nil, huma.Error400BadRequest("unknown template type")
	}
	tmpl, err := h.store.Upsert(ctx, input.ID, input.Type, input.Body.Subject, input.Body.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal error")
	}
	return &templateUpsertOutput{Body: tmpl}, nil
}
