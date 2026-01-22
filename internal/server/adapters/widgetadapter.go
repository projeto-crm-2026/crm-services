package adapters

import (
	"context"

	"github.com/projeto-crm-2026/crm-services/internal/server/middleware"
	"github.com/projeto-crm-2026/crm-services/internal/service/widgetservice"
)

type widgetValidatorAdapter struct {
	service widgetservice.WidgetService
}

func NewWidgetValidator(service widgetservice.WidgetService) middleware.APIKeyValidator {
	return &widgetValidatorAdapter{service: service}
}

func (a *widgetValidatorAdapter) ValidateAPIKey(ctx context.Context, publicKey, origin string) (*middleware.WidgetContext, error) {
	info, err := a.service.ValidateAPIKey(ctx, publicKey, origin)
	if err != nil {
		return nil, err
	}

	return &middleware.WidgetContext{
		UserID:    info.UserID,
		PublicKey: info.PublicKey,
		Domain:    info.Domain,
	}, nil
}
