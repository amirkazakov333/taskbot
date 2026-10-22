package reminder

import (
	"context"
	"log"
	"time"

	"telegram-bot/internal/model"
	"telegram-bot/internal/service"

	tgBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Worker struct {
	service  *service.ReminderService
	bot      *tgBot.Bot
	location *time.Location
}

func NewWorker(
	service *service.ReminderService,
	bot *tgBot.Bot,
) *Worker {

	location, err := time.LoadLocation("Asia/Dushanbe")
	if err != nil {
		location = time.FixedZone(
			"UTC+5",
			5*60*60,
		)
	}

	return &Worker{
		service:  service,
		bot:      bot,
		location: location,
	}
}

func (w *Worker) Start(
	ctx context.Context,
) {

	log.Println("Reminder worker started")

	// =========================
	// Проверка сразу после запуска
	// =========================

	now := time.Now().In(w.location)

	w.check(
		ctx,
		now,
		false, // не отправляем "через 30 минут"
		true,  // проверяем просроченные
	)

	// =========================
	// Дальше работаем по :00 / :30
	// =========================

	for {

		now := time.Now().In(w.location)

		nextRun := nextHalfHour(now)

		log.Printf(
			"next reminder check: %s",
			nextRun.Format("15:04:05"),
		)

		wait := time.Until(nextRun)

		timer := time.NewTimer(wait)

		select {

		case <-ctx.Done():

			if !timer.Stop() {
				<-timer.C
			}

			log.Println("Reminder worker stopped")
			return

		case <-timer.C:

			checkTime := time.Now().In(
				w.location,
			)

			// Каждые :00 проверяем overdue.
			// В :30 не проверяем.
			includeOverdue :=
				nextRun.Minute() == 0

			w.check(
				ctx,
				checkTime,
				true,
				includeOverdue,
			)
		}
	}
}

func nextHalfHour(
	now time.Time,
) time.Time {

	if now.Minute() < 30 {

		return time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			30,
			0,
			0,
			now.Location(),
		)
	}

	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour()+1,
		0,
		0,
		0,
		now.Location(),
	)
}

func (w *Worker) check(
	ctx context.Context,
	now time.Time,
	includePre bool,
	includeOverdue bool,
) {

	reminders, err := w.service.GetReminders(
		ctx,
		now,
		includePre,
		includeOverdue,
	)

	if err != nil {
		log.Printf(
			"get reminders error: %v",
			err,
		)
		return
	}

	for _, reminder := range reminders {

		err := w.send(
			ctx,
			reminder,
		)

		if err != nil {
			log.Printf(
				"send reminder error task_id=%d: %v",
				reminder.TaskID,
				err,
			)

			continue
		}

		// Только после успешной отправки
		// отмечаем уведомление в БД.
		if err := w.service.MarkSent(
			ctx,
			reminder,
			time.Now().In(w.location),
		); err != nil {

			log.Printf(
				"mark reminder sent error task_id=%d: %v",
				reminder.TaskID,
				err,
			)
		}
	}
}

func (w *Worker) send(
	ctx context.Context,
	reminder model.ReminderMessage,
) error {

	_, err := w.bot.SendMessage(
		ctx,
		&tgBot.SendMessageParams{
			ChatID: reminder.TelegramID,
			Text:   reminder.Text,

			ParseMode: models.ParseModeHTML,
		},
	)

	return err
}
