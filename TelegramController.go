package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/kneu-messenger-pigeon/authorizer-client"
	framework "github.com/kneu-messenger-pigeon/client-framework"
	"github.com/kneu-messenger-pigeon/client-framework/delayedDeleter/contracts"
	"github.com/kneu-messenger-pigeon/client-framework/models"
	"github.com/kneu-messenger-pigeon/events"
	scoreApi "github.com/kneu-messenger-pigeon/score-api"
	"github.com/kneu-messenger-pigeon/score-client"
	"gopkg.in/telebot.v3"
	tele "gopkg.in/telebot.v3"
	"io"
	"strconv"
	"strings"
	"sync"
)

const startCommand = "/start"

const listCommand = "/list"

const resetCommand = "/reset"

const TelegramControllerStartedMessage = "Telegram controller started\n"

type TelegramController struct {
	out                            io.Writer
	debugLogger                    *framework.DebugLogger
	bot                            *telebot.Bot
	composer                       framework.MessageComposerInterface
	userRepository                 framework.UserRepositoryInterface
	userLogoutHandler              framework.UserLogoutHandlerInterface
	authorizerClient               authorizer.ClientInterface
	scoreClient                    score.ClientInterface
	welcomeAnonymousDelayedDeleter contracts.DeleterInterface

	authRedirectUrl string

	markups struct {
		disciplineButton           *telebot.InlineButton
		listButton                 *telebot.InlineButton
		disciplineScoreReplyMarkup *telebot.ReplyMarkup
		authorizedUserReplyMarkup  *telebot.ReplyMarkup
		logoutUserReplyMarkup      *telebot.ReplyMarkup
	}
}

func NewTelegramController(serviceContainer *framework.ServiceContainer, bot *tele.Bot, out io.Writer) *TelegramController {
	return &TelegramController{
		out:                            out,
		debugLogger:                    serviceContainer.DebugLogger,
		bot:                            bot,
		composer:                       framework.NewMessageComposer(framework.MessageComposerConfig{}),
		userRepository:                 serviceContainer.UserRepository,
		userLogoutHandler:              serviceContainer.UserLogoutHandler,
		authorizerClient:               serviceContainer.AuthorizerClient,
		scoreClient:                    serviceContainer.ScoreClient,
		welcomeAnonymousDelayedDeleter: serviceContainer.WelcomeAnonymousDelayedDeleter,
	}
}

func (controller *TelegramController) Init() {
	controller.composer.SetPostFilter(escapeMarkDown)
	controller.authRedirectUrl = fmt.Sprintf("https://t.me/%s?start", controller.bot.Me.Username)

	controller.markups.disciplineButton = &telebot.InlineButton{
		Unique: "discipline",
	}
	controller.markups.listButton = &telebot.InlineButton{
		Text:   "Назад",
		Unique: "list",
	}

	controller.markups.disciplineScoreReplyMarkup = &tele.ReplyMarkup{
		OneTimeKeyboard: true,
		InlineKeyboard: [][]tele.InlineButton{
			{*controller.markups.listButton},
		},
	}

	controller.markups.authorizedUserReplyMarkup = &telebot.ReplyMarkup{
		ResizeKeyboard: true,
		ReplyKeyboard: [][]tele.ReplyButton{
			{
				{Text: listCommand + " Мої результати"},
			},
		},
	}

	controller.markups.logoutUserReplyMarkup = &telebot.ReplyMarkup{
		ResizeKeyboard: true,
		ReplyKeyboard: [][]tele.ReplyButton{
			{
				{Text: startCommand + " Запустити!"},
			},
		},
	}

	controller.setupRoutes()
}

func (controller *TelegramController) Execute(ctx context.Context, wg *sync.WaitGroup) {
	controller.Init()

	go controller.bot.Start()
	_, _ = fmt.Fprint(controller.out, TelegramControllerStartedMessage)
	<-ctx.Done()
	controller.bot.Stop()
	wg.Done()
}

func (controller *TelegramController) setupRoutes() {
	controller.bot.Use(onlyPrivateChatMiddleware())
	controller.bot.Use(authMiddleware(controller.userRepository))
	controller.bot.Use(onlyAuthorizedMiddleware(controller.WelcomeAnonymousAction))

	controller.bot.Handle(resetCommand, controller.ResetAction)
	controller.bot.Handle(startCommand, controller.DisciplinesListAction)
	controller.bot.Handle(listCommand, controller.DisciplinesListAction)
	controller.bot.Handle(controller.markups.listButton, controller.DisciplinesListAction)
	controller.bot.Handle(controller.markups.disciplineButton, controller.DisciplineScoresAction)
	controller.bot.Handle(tele.OnText, controller.DisciplinesListAction)
}

func (controller *TelegramController) ResetAction(c tele.Context) error {
	return controller.userLogoutHandler.Handle(strconv.FormatInt(c.Chat().ID, 10))
}

func (controller *TelegramController) WelcomeAnonymousAction(c tele.Context) error {
	authUrl, expireAt, err := controller.authorizerClient.GetAuthUrl(
		strconv.FormatInt(c.Chat().ID, 10),
		controller.authRedirectUrl,
	)

	if err != nil {
		_, _ = fmt.Fprintf(controller.out, "failed to get Auth url: %v\n", err)
		return err
	}

	err, messageText := controller.composer.ComposeWelcomeAnonymousMessage(
		models.WelcomeAnonymousMessageData{
			AuthUrl:  authUrl,
			ExpireAt: expireAt,
		},
	)
	if err != nil {
		return err
	}

	var message *tele.Message
	message, err = controller.bot.Send(c.Recipient(), messageText, tele.Protected, controller.markups.logoutUserReplyMarkup)

	if err != nil {
		return err
	}

	controller.welcomeAnonymousDelayedDeleter.AddToQueue(&contracts.DeleteTask{
		ScheduledAt: expireAt.Unix(),
		MessageId:   int32(message.ID),
		ChatId:      c.Chat().ID,
	})

	return nil
}

func (controller *TelegramController) HandleDeleteTask(task *contracts.DeleteTask) error {
	return controller.bot.Delete(tele.StoredMessage{
		MessageID: strconv.Itoa(int(task.GetMessageId())),
		ChatID:    task.GetChatId(),
	})
}

func (controller *TelegramController) WelcomeAuthorizedAction(event *events.UserAuthorizedEvent) error {
	student := controller.userRepository.GetStudent(event.ClientUserId)

	err, message := controller.composer.ComposeWelcomeAuthorizedMessage(
		models.UserAuthorizedMessageData{
			StudentMessageData: models.NewStudentMessageData(student),
		},
	)
	if err == nil {
		_, err = controller.bot.Send(
			makeChatId(event.ClientUserId),
			message,
			controller.markups.authorizedUserReplyMarkup,
		)

		if err != nil {
			_, _ = fmt.Fprintf(controller.out, "WelcomeAuthorizedAction failed to send message: %v; text: %s\n", err, message)
		}
	}

	return err
}

func (controller *TelegramController) LogoutFinishedAction(event *events.UserAuthorizedEvent) error {
	err, message := controller.composer.ComposeLogoutFinishedMessage()
	if err == nil {
		_, err = controller.bot.Send(makeChatId(event.ClientUserId), message, controller.markups.logoutUserReplyMarkup)

		if err != nil && !isBlockedByUserErr(err) {
			_, _ = fmt.Fprintf(controller.out, "LogoutFinishedAction failed to send message: %v; text: %s\n", err, message)
		}

	}
	return err
}

func (controller *TelegramController) DisciplinesListAction(c tele.Context) error {
	DisciplinesListActionRequestTotal.Inc()

	student := getStudent(c)

	disciplines, err := controller.scoreClient.GetStudentDisciplines(student.Id)
	if err == nil {
		replyMarkup := &tele.ReplyMarkup{
			OneTimeKeyboard: true,
			InlineKeyboard:  make([][]tele.InlineButton, len(disciplines)),
		}

		var disciplineButton *tele.InlineButton

		for i, discipline := range disciplines {
			disciplineButton = controller.markups.disciplineButton.With(strconv.Itoa(discipline.Discipline.Id))
			disciplineButton.Text = discipline.Discipline.Name

			replyMarkup.InlineKeyboard[i] = []tele.InlineButton{
				*disciplineButton,
			}
		}

		var message string
		err, message = controller.composer.ComposeDisciplinesListMessage(
			models.DisciplinesListMessageData{
				StudentMessageData: models.NewStudentMessageData(student),
				Disciplines:        disciplines,
			},
		)
		if err == nil {
			err = c.Send(message, replyMarkup)
		}
	}

	return err
}

func (controller *TelegramController) DisciplineScoresAction(c tele.Context) error {
	DisciplineScoresActionRequestTotal.Inc()

	var message string
	student := getStudent(c)
	disciplineId, _ := strconv.Atoi(c.Callback().Data)

	discipline, err := controller.scoreClient.GetStudentDiscipline(student.Id, disciplineId)

	if err != nil {
		controller.removeReplyMarkup(c.Message())
	} else {
		err, message = controller.composer.ComposeDisciplineScoresMessage(
			models.DisciplinesScoresMessageData{
				StudentMessageData: models.NewStudentMessageData(student),
				Discipline:         discipline,
			},
		)

		if err == nil {
			err = c.Send(message, controller.markups.disciplineScoreReplyMarkup)
		}
	}

	if err != nil && strings.Contains(err.Error(), "Bad Request: can't parse entities") {
		_, _ = fmt.Fprintf(controller.out, "DisciplineScoresAction failed to send message: %v; text: %s\n", err, message)
	}

	return err
}

func (controller *TelegramController) ScoreChangedAction(
	chatId string, previousMessageId string,
	disciplineScore *scoreApi.DisciplineScore, previousScore *scoreApi.Score,
) (err error, messageId string) {
	messageData := models.ScoreChangedMessageData{
		Discipline: disciplineScore.Discipline,
		Score:      disciplineScore.Score,
		Previous:   *previousScore,
	}

	err, messageText := controller.composer.ComposeScoreChanged(messageData)
	if err == nil {
		disciplineButton := controller.markups.disciplineButton.With(strconv.Itoa(disciplineScore.Discipline.Id))
		disciplineButton.Text = disciplineScore.Discipline.Name

		replyMarkup := &tele.ReplyMarkup{
			OneTimeKeyboard: true,
			InlineKeyboard: [][]tele.InlineButton{
				{
					*disciplineButton,
				},
			},
		}

		chatIdInt64 := makeInt64(chatId)
		var message *tele.Message
		if disciplineScore.Score.IsEqual(previousScore) {
			if previousMessageId != "" {
				err = controller.bot.Delete(tele.StoredMessage{
					MessageID: previousMessageId,
					ChatID:    chatIdInt64,
				})
				controller.debugLogger.Log(
					"ScoreChangedAction: delete message with id %s, chatId %s; err: %v",
					previousMessageId, chatId, err,
				)
			}

		} else if previousMessageId == "" {
			message, err = controller.bot.Send(tele.ChatID(chatIdInt64), messageText, replyMarkup)
			controller.debugLogger.Log(
				"ScoreChangedAction: send new message to %s; err: %v; message: %#v",
				chatId, err, message,
			)

		} else {
			message, err = controller.bot.Edit(tele.StoredMessage{
				MessageID: previousMessageId,
				ChatID:    chatIdInt64,
			}, messageText, replyMarkup)

			controller.debugLogger.Log(
				"ScoreChangedAction: edit message with id %s, chatId %s; err: %v; message: %#v",
				previousMessageId, chatId, err, message,
			)

			if errors.Is(err, tele.ErrSameMessageContent) || errors.Is(err, tele.ErrMessageNotModified) {
				_, _ = fmt.Fprintln(controller.out, `Ignore error "message not modified"`)
				return nil, previousMessageId
			}
		}

		err = controller.handleTelegramError(err, chatIdInt64)

		if message != nil {
			return err, strconv.Itoa(message.ID)
		}
	}

	return err, ""
}

func (controller *TelegramController) handleTelegramError(err error, chatId int64) error {
	if isBlockedByUserErr(err) {
		// rewrite error to result of userLogoutHandler.Handle
		return controller.userLogoutHandler.Handle(strconv.FormatInt(chatId, 10))
	}

	return err
}

func (controller *TelegramController) removeReplyMarkup(message tele.Editable) {
	if message != nil {
		_, err := controller.bot.EditReplyMarkup(message, nil)
		if err != nil {
			_, _ = fmt.Fprintf(controller.out, "Failed to remove reply markup: %v\n", err)
		}
	}
}
