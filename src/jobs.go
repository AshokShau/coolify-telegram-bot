package src

import (
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"

	td "github.com/AshokShau/gotdbot"
)

const pageSize = 5

func jobsHandler(c *td.Client, msg *td.Message) error {
	text, kb, err := buildJobsMessage(1)
	if err != nil {
		_, err = msg.ReplyText(c, err.Error(), sendOpts(msg, nil))
		return err
	}

	_, err = msg.ReplyText(c, text, sendOpts(msg, kb))
	return err
}

func jobsPaginationHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	_ = cb.Answer(c, 0, false, "Processing...", "")

	data := cb.DataString()
	page := 1
	if parts := strings.Split(data, ":"); len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &page)
	}

	text, kb, err := buildJobsMessage(page)
	if err != nil {
		return editCallback(c, cb, err.Error(), &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func unscheduleCallbackHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	taskID := strings.TrimPrefix(cb.DataString(), "unschedule:")

	if err := scheduler.RemoveTask(taskID); err != nil {
		_ = cb.Answer(c, 0, true, fmt.Sprintf("Warning: %v", err), "")
	} else {
		_ = cb.Answer(c, 0, false, "Task removed from scheduler.", "")
	}

	if err := database.DeleteTask(taskID); err != nil {
		_ = cb.Answer(c, 0, true, fmt.Sprintf("Error deleting task: %v", err), "")
		return err
	}

	text, kb, err := buildJobsMessage(1)
	if err != nil {
		return editCallback(c, cb, "Task deleted. No scheduled jobs remaining.", &td.EditTextMessageOpts{ReplyMarkup: makeBackButton("start_menu")})
	}

	return editCallback(c, cb, text, &td.EditTextMessageOpts{ReplyMarkup: kb})
}

func buildJobsMessage(page int) (string, td.ReplyMarkup, error) {
	tasks, err := database.GetTasks()
	if err != nil {
		return "", nil, fmt.Errorf("error fetching tasks: %v", err)
	}

	if len(tasks) == 0 {
		kb := makeBackButton("start_menu")
		return "No scheduled jobs found.", kb, nil
	}

	start, end, buttons := Paginate(len(tasks), page, pageSize, "jobs:")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Scheduled Jobs (Page %d):</b>\n\n", page))

	kb := &td.ReplyMarkupInlineKeyboard{}

	for _, task := range tasks[start:end] {
		hexID := task.ID.Hex()
		shortID := hexID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		sb.WriteString(fmt.Sprintf("ID: <code>%s</code>\n", hexID))
		sb.WriteString(fmt.Sprintf("Name: %s\n", task.Name))
		sb.WriteString(fmt.Sprintf("Type: %s\n", task.Type))
		sb.WriteString(fmt.Sprintf("Schedule: %s\n", task.Schedule))
		if task.OneTime {
			sb.WriteString(fmt.Sprintf("Next Run: %s\n", task.NextRun.Format("2006-01-02 15:04:05")))
		}
		sb.WriteString("--------------------\n")

		btnText := fmt.Sprintf("Unschedule %s (%s)", task.Name, shortID)
		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			makeCallbackButton(btnText, "unschedule:"+hexID),
		})
	}

	if row := buildPaginationButtonsRow(buttons); len(row) > 0 {
		kb.Rows = append(kb.Rows, row)
	}

	kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
		makeCallbackButton("Main Menu", "start_menu"),
	})

	return sb.String(), kb, nil
}
