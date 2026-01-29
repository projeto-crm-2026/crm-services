package widgetservice

import "errors"

var (
	ErrAPIKeyInactive   = errors.New("API key is inactive")
	ErrOriginNotAllowed = errors.New("origin not allowed for this API key")
	ErrAPIKeyNotFound   = errors.New("API key not found")
)
