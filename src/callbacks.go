package src

import (
	"coolifymanager/src/config"
	"fmt"
	"html"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func listProjectsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}

	cb := ctx.CallbackQuery
	_, _ = cb.Answer(b, nil)
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Failed to fetch projects: "+err.Error(), nil)
		return err
	}

	if len(apps) == 0 {
		_, _, err = cb.Message.EditText(b, "😶 No applications found.", nil)
		return err
	}

	var buttons [][]gotgbot.InlineKeyboardButton

	for _, app := range apps {
		text := fmt.Sprintf("📦 %s (%s)", app.Name, app.Status)
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: text, CallbackData: "project_menu:" + app.UUID},
		})
	}

	_, _, err = cb.Message.EditText(b, "<b>📋 Select a project:</b>", &gotgbot.EditMessageTextOpts{
		ParseMode:   "HTML",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: buttons},
	})
	return err
}

func projectMenuHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}

	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "project_menu:")

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Failed to load project: "+err.Error(), nil)
		return err
	}

	text := fmt.Sprintf("<b>📦 %s</b>\n🌐 %s\n📄 Status: <code>%s</code>", app.Name, app.FQDN, app.Status)
	btns := [][]gotgbot.InlineKeyboardButton{
		{{Text: "🔄 Restart", CallbackData: "restart:" + uuid}, {Text: "🚀 Deploy", CallbackData: "deploy:" + uuid}},
		{{Text: "📜 Logs", CallbackData: "logs:" + uuid}, {Text: "ℹ️ Status", CallbackData: "status:" + uuid}},
		{{Text: "🛑 Stop", CallbackData: "stop:" + uuid}, {Text: "❌ Delete", CallbackData: "delete:" + uuid}},
		{{Text: "🔙 Back", CallbackData: "list_projects:"}},
	}

	_, _, err = cb.Message.EditText(b, text, &gotgbot.EditMessageTextOpts{
		ParseMode:   "HTML",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: btns},
	})
	return err
}

func restartHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}
	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "restart:")
	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Restart failed: "+err.Error(), nil)
		return err
	}
	text := fmt.Sprintf("✅ Restart queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, _, err = cb.Message.EditText(b, text, &gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
	return err
}

func deployHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}
	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "deploy:")
	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Deploy failed: "+err.Error(), nil)
		return err
	}
	text := fmt.Sprintf("✅ Deployment queued!\nDeployment UUID: <code>%s</code>", res.DeploymentUUID)
	_, _, err = cb.Message.EditText(b, text, &gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
	return err
}

func logsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}
	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "logs:")
	logs, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_, _, _ = cb.Message.EditText(b, "❌ Logs error: "+err.Error(), nil)
		return ext.EndGroups
	}

	_, _, err = cb.Message.EditText(b, "<b>📜 Logs</b>\n"+html.EscapeString(logs), &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})

	return err
}

func statusHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}
	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "status:")
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Status error: "+err.Error(), nil)
		return nil
	}

	text := fmt.Sprintf("📦 <b>%s</b>\nCurrent Status: <code>%s</code>", app.Name, app.Status)
	_, _, err = cb.Message.EditText(b, text, &gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
	return err
}

func stopHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}
	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "stop:")
	res, err := config.Coolify.StopApplicationByUUID(uuid)
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Stop failed: "+err.Error(), nil)
		return nil
	}

	_, _, err = cb.Message.EditText(b, "🛑 "+res.Message, nil)
	return err
}

func deleteHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if !config.IsDev(ctx.EffectiveUser.Id) {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "🚫 You are not authorized.",
			ShowAlert: true,
		})
		return nil
	}
	_, _ = cb.Answer(b, nil)

	uuid := strings.TrimPrefix(cb.Data, "delete:")
	err := config.Coolify.DeleteApplicationByUUID(uuid)
	if err != nil {
		_, _, err = cb.Message.EditText(b, "❌ Delete failed: "+err.Error(), nil)
		return nil
	}

	_, _, err = cb.Message.EditText(b, "✅ Application deleted successfully.", nil)
	return err
}
