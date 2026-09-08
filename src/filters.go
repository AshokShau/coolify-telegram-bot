package src

import (
	"strings"

	"coolifymanager/src/config"

	td "github.com/AshokShau/gotdbot"
)

// IsDevMsg checks if the message sender is an authorized developer
func IsDevMsg(m *td.Message) bool {
	return config.IsDev(m.SenderID())
}

// DevCallbackPrefix returns a filter function for callback queries checking developer status and prefix match
func DevCallbackPrefix(prefix string) func(*td.UpdateNewCallbackQuery) bool {
	return func(cb *td.UpdateNewCallbackQuery) bool {
		return config.IsDev(cb.SenderUserId) && strings.HasPrefix(cb.DataString(), prefix)
	}
}
