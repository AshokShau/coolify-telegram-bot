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
func DevCallbackPrefix(c *td.Client, prefix string) func(*td.UpdateNewCallbackQuery) bool {
	return func(cb *td.UpdateNewCallbackQuery) bool {
		if !strings.HasPrefix(cb.DataString(), prefix) {
			return false
		}
		if !config.IsDev(cb.SenderUserId) {
			_ = cb.Answer(c, 0, true, "You are not authorized.", "")
			return false
		}
		return true
	}
}
