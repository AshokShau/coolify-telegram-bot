package src

import (
	td "github.com/AshokShau/gotdbot"
)

func makeCallbackButton(text, data string) td.InlineKeyboardButton {
	return td.InlineKeyboardButton{
		Text: text,
		Type: &td.InlineKeyboardButtonTypeCallback{
			Data: []byte(data),
		},
	}
}

func buildPaginationButtonsRow(buttons []PageButton) []td.InlineKeyboardButton {
	if len(buttons) == 0 {
		return nil
	}
	row := make([]td.InlineKeyboardButton, 0, len(buttons))
	for _, btn := range buttons {
		row = append(row, makeCallbackButton(btn.Text, btn.Data))
	}
	return row
}

func makeBackButton(data string) *td.ReplyMarkupInlineKeyboard {
	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				makeCallbackButton("Back", data),
			},
		},
	}
}

func sendOpts(msg *td.Message, replyMarkup td.ReplyMarkup) *td.SendTextMessageOpts {
	opts := &td.SendTextMessageOpts{ReplyMarkup: replyMarkup}
	if !msg.IsPrivate() {
		opts.ReceiverUserID = msg.SenderID()
	}
	return opts
}

func docOpts(msg *td.Message, caption string, replyMarkup td.ReplyMarkup) *td.SendDocumentOpts {
	opts := &td.SendDocumentOpts{
		Caption:     caption,
		ReplyMarkup: replyMarkup,
	}
	if !msg.IsPrivate() {
		opts.ReceiverUserID = msg.SenderID()
	}
	return opts
}

func makeEphemeralOpts(opts *td.EditTextMessageOpts) *td.EditEphemeralMessageTextOpts {
	if opts == nil {
		return &td.EditEphemeralMessageTextOpts{}
	}
	return &td.EditEphemeralMessageTextOpts{
		ParseMode:             opts.ParseMode,
		DisableWebPagePreview: opts.DisableWebPagePreview,
		Url:                   opts.Url,
		ForceSmallMedia:       opts.ForceSmallMedia,
		ForceLargeMedia:       opts.ForceLargeMedia,
		ShowAboveText:         opts.ShowAboveText,
		ReplyMarkup:           opts.ReplyMarkup,
	}
}

func editMsg(c *td.Client, msg *td.Message, text string, opts *td.EditTextMessageOpts) (*td.Message, error) {
	if !msg.IsPrivate() {
		eOpts := makeEphemeralOpts(opts)
		if err := c.EditEphemeralMessageText(msg.ChatId, msg.EphemeralMessageId, msg.SenderID(), text, eOpts); err == nil {
			return msg, nil
		}
	}
	return msg.EditText(c, text, opts)
}

func editCallback(c *td.Client, cb *td.UpdateNewCallbackQuery, text string, opts *td.EditTextMessageOpts) error {
	if !cb.IsPrivate() {
		msg, err := cb.GetMessage(c)
		if err != nil {
			return err
		}
		eOpts := makeEphemeralOpts(opts)
		return c.EditEphemeralMessageText(cb.ChatId, msg.EphemeralMessageId, cb.SenderUserId, text, eOpts)
	}

	_, err := cb.EditMessageText(c, text, opts)
	return err
}
