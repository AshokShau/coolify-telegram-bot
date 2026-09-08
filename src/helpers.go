package src

import (
	td "github.com/AshokShau/gotdbot"
)

// replyMsg replies to a message, automatically applying ReceiverUserID for group ephemeral responses
func replyMsg(c *td.Client, msg *td.Message, text string, opts *td.SendTextMessageOpts) (*td.Message, error) {
	if opts == nil {
		opts = &td.SendTextMessageOpts{}
	}
	if opts.ParseMode == "" {
		opts.ParseMode = "HTML"
	}
	if !msg.IsPrivate() {
		opts.ReceiverUserID = msg.SenderID()
	}
	return msg.ReplyText(c, text, opts)
}

// replyDoc replies with a document, applying ReceiverUserID if in a group
func replyDoc(c *td.Client, msg *td.Message, path string, caption string, replyMarkup td.ReplyMarkup) (*td.Message, error) {
	opts := &td.SendDocumentOpts{
		Caption:     caption,
		ParseMode:   "HTML",
		ReplyMarkup: replyMarkup,
	}
	if !msg.IsPrivate() {
		opts.ReceiverUserID = msg.SenderID()
	}
	return msg.ReplyDocument(c, td.InputFileLocal{Path: path}, opts)
}

// editMsg edits an existing message, supporting both normal and ephemeral message edits
func editMsg(c *td.Client, msg *td.Message, text string, opts *td.EditTextMessageOpts) (*td.Message, error) {
	if opts == nil {
		opts = &td.EditTextMessageOpts{}
	}
	if opts.ParseMode == "" {
		opts.ParseMode = "HTML"
	}

	if !msg.IsPrivate() {
		eOpts := &td.EditEphemeralMessageTextOpts{
			ParseMode:             opts.ParseMode,
			DisableWebPagePreview: opts.DisableWebPagePreview,
			Url:                   opts.Url,
			ForceSmallMedia:       opts.ForceSmallMedia,
			ForceLargeMedia:       opts.ForceLargeMedia,
			ShowAboveText:         opts.ShowAboveText,
			ReplyMarkup:           opts.ReplyMarkup,
		}
		err := c.EditEphemeralMessageText(msg.ChatId, int32(msg.Id), msg.SenderID(), text, eOpts)
		if err == nil {
			return msg, nil
		}
	}

	return msg.EditText(c, text, opts)
}

// editCallback edits the message associated with a callback query
func editCallback(c *td.Client, cb *td.UpdateNewCallbackQuery, text string, opts *td.EditTextMessageOpts) error {
	if opts == nil {
		opts = &td.EditTextMessageOpts{}
	}
	if opts.ParseMode == "" {
		opts.ParseMode = "HTML"
	}

	if !cb.IsPrivate() {
		eOpts := &td.EditEphemeralMessageTextOpts{
			ParseMode:             opts.ParseMode,
			DisableWebPagePreview: opts.DisableWebPagePreview,
			Url:                   opts.Url,
			ForceSmallMedia:       opts.ForceSmallMedia,
			ForceLargeMedia:       opts.ForceLargeMedia,
			ShowAboveText:         opts.ShowAboveText,
			ReplyMarkup:           opts.ReplyMarkup,
		}
		err := c.EditEphemeralMessageText(cb.ChatId, int32(cb.MessageId), cb.SenderUserId, text, eOpts)
		if err == nil {
			return nil
		}
	}

	_, err := cb.EditMessageText(c, text, opts)
	return err
}

// makeBackButton creates a standard inline keyboard with a single Back button
func makeBackButton(data string) *td.ReplyMarkupInlineKeyboard {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "Back",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(data),
					},
				},
			},
		},
	}
}
