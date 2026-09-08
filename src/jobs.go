package src

import (
	"coolifymanager/src/database"
	"fmt"
	"strings"

	td "github.com/AshokShau/gotdbot"
)

const pageSize = 5

func jobsHandler(c *td.Client, msg *td.Message) error {
	text, kb, err := buildJobsMessage(1)
	if err != nil {
		_, err = replyMsg(c, msg, err.Error(), nil)
		return err
	}

	_, err = replyMsg(c, msg, text, &td.SendTextMessageOpts{ReplyMarkup: kb})
	return err
}

func jobsPaginationHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	data := cb.DataString()
	page := 1
	if parts := strings.Split(data, ":"); len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &page)
	}

	text, kb, err := buildJobsMessage(page)
	if err != nil {
		_ = cb.Answer(c, 0, true, err.Error(), "")
		return nil
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func buildJobsMessage(page int) (string, td.ReplyMarkup, error) {
	tasks, err := database.GetTasks()
	if err != nil {
		return "", nil, fmt.Errorf("error fetching tasks: %v", err)
	}

	if len(tasks) == 0 {
		return "No scheduled jobs found.", nil, nil
	}

	start, end, buttons := Paginate(len(tasks), page, pageSize, "jobs:")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Scheduled Jobs (Page %d):</b>\n\n", page))

	for _, task := range tasks[start:end] {
		sb.WriteString(fmt.Sprintf("ID: <code>%s</code>\n", task.ID.Hex()))
		sb.WriteString(fmt.Sprintf("Name: %s\n", task.Name))
		sb.WriteString(fmt.Sprintf("Type: %s\n", task.Type))
		sb.WriteString(fmt.Sprintf("Schedule: %s\n", task.Schedule))
		if task.OneTime {
			sb.WriteString(fmt.Sprintf("Next Run: %s\n", task.NextRun.Format("2006-01-02 15:04:05")))
		}
		sb.WriteString("--------------------\n")
	}

	kb := &td.ReplyMarkupInlineKeyboard{}
	if len(buttons) > 0 {
		row := make([]td.InlineKeyboardButton, 0, len(buttons))

		for _, btn := range buttons {
			row = append(row, td.InlineKeyboardButton{
				Text: btn.Text,
				Type: &td.InlineKeyboardButtonTypeCallback{
					Data: []byte(btn.Data),
				},
			})
		}

		kb.Rows = append(kb.Rows, row)
	}

	return sb.String(), kb, nil
}
