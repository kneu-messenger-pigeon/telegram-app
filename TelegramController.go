package main

import (
	"context"
	"fmt"
	"github.com/kneu-messenger-pigeon/authorizer-client"
	framework "github.com/kneu-messenger-pigeon/client-framework"
	"github.com/kneu-messenger-pigeon/client-framework/models"
	"github.com/kneu-messenger-pigeon/events"
	"github.com/kneu-messenger-pigeon/score-client"
	"gopkg.in/telebot.v3"
	tele "gopkg.in/telebot.v3"
	"io"
	"strconv"
	"sync"
)

const listCommand = "/list"

type TelegramController struct {
	out               io.Writer
	bot               *telebot.Bot
	composer          framework.MessageComposerInterface
	userRepository    *framework.UserRepository
	userLogoutHandler framework.UserLogoutHandlerInterface
	authorizerClient  *authorizer.Client
	scoreClient       score.ClientInterface

	authRedirectUrl string

	markups struct {
		disciplineButton           *telebot.InlineButton
		listButton                 *telebot.InlineButton
		disciplineScoreReplyMarkup *telebot.ReplyMarkup
		authorizedUserReplyMarkup  *telebot.ReplyMarkup
		logoutUserReplyMarkup      *telebot.ReplyMarkup
	}
}

func (controller *TelegramController) Init() {
	controller.authRedirectUrl = fmt.Sprintf("https://t.me/%s?start", controller.bot.Me.Username)

	controller.markups.disciplineButton = &telebot.InlineButton{
		Unique: "discipline",
	}
	controller.markups.listButton = &telebot.InlineButton{
		Text:   "Назад",
		Unique: "list",
	}

	controller.markups.disciplineScoreReplyMarkup = &tele.ReplyMarkup{
		InlineKeyboard: [][]tele.InlineButton{
			{*controller.markups.listButton},
		},
	}

	controller.markups.authorizedUserReplyMarkup = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{
			{
				{Text: listCommand + " Мої результати"},
			},
		},
	}

	controller.markups.logoutUserReplyMarkup = &telebot.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{
			{
				{Text: "/start Запустити!"},
			},
		},
	}

	controller.setupRoutes()
}

func (controller *TelegramController) Execute(ctx context.Context, wg *sync.WaitGroup) {
	controller.Init()

	go controller.bot.Start()
	<-ctx.Done()
	controller.bot.Stop()
	wg.Done()
}

func (controller *TelegramController) setupRoutes() {
	controller.bot.Use(onlyPrivateChatMiddleware())
	controller.bot.Use(authMiddleware(controller.userRepository))
	controller.bot.Use(onlyAuthorizedMiddleware(controller.WelcomeAnonymousAction))

	controller.bot.Handle("/reset", controller.ResetAction)
	controller.bot.Handle("/start", controller.DisciplinesListAction)
	controller.bot.Handle(listCommand, controller.DisciplinesListAction)
	controller.bot.Handle(controller.markups.listButton, controller.DisciplinesListAction)
	controller.bot.Handle(controller.markups.disciplineButton, controller.DisciplineScoresAction)
	controller.bot.Handle(tele.OnText, controller.DisciplinesListAction)
}

func (controller *TelegramController) ResetAction(c tele.Context) error {
	return controller.userLogoutHandler.Handle(strconv.FormatInt(c.Chat().ID, 10))
}

func (controller *TelegramController) WelcomeAnonymousAction(c tele.Context) error {
	authUrl, err := controller.authorizerClient.GetAuthUrl(
		strconv.FormatInt(c.Chat().ID, 10),
		controller.authRedirectUrl,
	)

	if err != nil {
		_, _ = fmt.Fprintf(controller.out, "failed to get Auth url: %v\n", err)
		return err
	}

	err, message := controller.composer.ComposeWelcomeAnonymousMessage(authUrl)
	if err == nil {
		err = c.Send(message)
	}
	return err
}

func (controller *TelegramController) WelcomeAuthorizedAction(event *events.UserAuthorizedEvent) error {
	student := controller.userRepository.GetStudent(event.ClientUserId)

	err, message := controller.composer.ComposeWelcomeAuthorizedMessage(
		models.UserAuthorizedMessageData{
			StudentMessageData: models.NewStudentMessageData(student),
		},
	)
	if err == nil {
		_, err = controller.bot.Send(makeChatId(event.ClientUserId), message, controller.markups.authorizedUserReplyMarkup)
	}

	return err
}

func (controller *TelegramController) LogoutFinishedAction(event *events.UserAuthorizedEvent) error {
	err, message := controller.composer.ComposeLogoutFinishedMessage()
	if err == nil {
		_, err = controller.bot.Send(makeChatId(event.ClientUserId), message, controller.markups.logoutUserReplyMarkup)
	}
	return err
}

func (controller *TelegramController) DisciplinesListAction(c tele.Context) error {
	student := getStudent(c)

	disciplines, err := controller.scoreClient.GetStudentDisciplines(student.Id)
	if err == nil {
		replyMarkup := &tele.ReplyMarkup{
			InlineKeyboard: make([][]tele.InlineButton, len(disciplines)),
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
			err = c.Send(message, &telebot.SendOptions{
				DisableWebPagePreview: true,
			}, replyMarkup)
		}
	}

	return err
}

func (controller *TelegramController) DisciplineScoresAction(c tele.Context) error {
	var message string
	student := getStudent(c)
	disciplineId, _ := strconv.Atoi(c.Callback().Data)

	discipline, err := controller.scoreClient.GetStudentDiscipline(student.Id, disciplineId)

	if err == nil {
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

	return err
}

func (controller *TelegramController) ScoreChangedAction(event *events.ScoreChangedEvent) error {
	chatIds := controller.userRepository.GetClientUserIds(event.StudentId)

	for _, chatId := range chatIds {
		_, _ = controller.bot.Send(makeChatId(chatId), "S")
	}

	return nil
}
