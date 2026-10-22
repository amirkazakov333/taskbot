package reminder

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

func taskActionKeyboard(
	taskID int64,
) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text: "✅ Выполнено",
					CallbackData: fmt.Sprintf(
						"task:done:%d",
						taskID,
					),
				},
			},
			{
				{
					Text: "❌ Отменить задачу",
					CallbackData: fmt.Sprintf(
						"task:cancel:%d",
						taskID,
					),
				},
			},
			{
				{
					Text: "🔄 Перенести задачу",
					CallbackData: fmt.Sprintf(
						"task:reschedule:%d",
						taskID,
					),
				},
			},
		},
	}
}
