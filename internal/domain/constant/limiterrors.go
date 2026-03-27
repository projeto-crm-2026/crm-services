package constant

import "errors"

var (
	ErrContactLimitReached       = errors.New("contact limit reached for your current plan")
	ErrMemberLimitReached        = errors.New("member limit reached for your current plan")
	ErrChatResponderLimitReached = errors.New("chat responder limit reached for your current plan")
)
