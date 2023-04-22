package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kneu-messenger-pigeon/authorizer-client"
	"github.com/kneu-messenger-pigeon/events"
	"github.com/kneu-messenger-pigeon/score-client"
	"gopkg.in/telebot.v3"
	tele "gopkg.in/telebot.v3"
	"html/template"
	"io"
	"strconv"
	"sync"
)

const listCommand = "/list"

type TelegramController struct {
	out                      io.Writer
	bot                      *telebot.Bot
	userRepository           *UserRepository
	authorizerClient         *authorizer.Client
	userLogoutHandler        UserLogoutHandlerInterface
	scoreClient              score.ClientInterface
	userAuthorizedEventQueue <-chan *events.UserAuthorizedEvent
	scoreChangedEventQueue   <-chan *events.ScoreChangedEvent
	templates                *template.Template
	authRedirectUrl          string

	markups struct {
		disciplineButton           *telebot.InlineButton
		listButton                 *telebot.InlineButton
		disciplineScoreReplyMarkup *telebot.ReplyMarkup
		authorizedUserReplyMarkup  *telebot.ReplyMarkup
		logoutUserReplyMarkup      *telebot.ReplyMarkup
	}
}

func (controller *TelegramController) Init() {
	controller.templates = template.Must(
		template.New("").
			Funcs(TemplateFunctionMap).
			ParseGlob("templates/*.html"),
	)

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

	controller.authRedirectUrl = fmt.Sprintf("https://t.me/%s?start", controller.bot.Me.Username)

	controller.setupRoutes()
}

func (controller *TelegramController) Execute(ctx context.Context, wg *sync.WaitGroup) {
	controller.Init()

	go controller.handleEvents(ctx)
	go controller.bot.Start()
	<-ctx.Done()
	controller.bot.Stop()
	wg.Done()
}

func (controller *TelegramController) setupRoutes() {
	controller.bot.Use(onlyPrivateChatMiddleware())
	controller.bot.Use(authMiddleware(controller.userRepository))
	controller.bot.Use(onlyAuthorizedMiddleware(controller.welcomeAnonymousAction))

	controller.bot.Handle("/reset", controller.resetAction)
	controller.bot.Handle("/start", controller.disciplinesListAction)
	controller.bot.Handle(listCommand, controller.disciplinesListAction)
	controller.bot.Handle(controller.markups.listButton, controller.disciplinesListAction)
	controller.bot.Handle(controller.markups.disciplineButton, controller.disciplineScoresAction)
	controller.bot.Handle(tele.OnText, controller.disciplinesListAction)
}

func (controller *TelegramController) handleEvents(ctx context.Context) {
	var userAuthorizedEvent *events.UserAuthorizedEvent
	var scoreChangeEvent *events.ScoreChangedEvent

	for {
		select {
		case userAuthorizedEvent = <-controller.userAuthorizedEventQueue:
			controller.userAuthorizedAction(userAuthorizedEvent)

		case scoreChangeEvent = <-controller.scoreChangedEventQueue:
			controller.scoreChangedAction(scoreChangeEvent)

		case <-ctx.Done():
			return
		}
	}
}

func (controller *TelegramController) resetAction(c tele.Context) error {
	return controller.userLogoutHandler.handle(strconv.FormatInt(c.Chat().ID, 10))
}

func (controller *TelegramController) userAuthorizedAction(event *events.UserAuthorizedEvent) {
	if event.StudentId != 0 {
		controller.welcomeAuthorizedAction(event)
	} else {
		controller.logoutFinishedAction(event)
	}
}

func (controller *TelegramController) welcomeAnonymousAction(c tele.Context) error {
	authUrl, err := controller.authorizerClient.GetAuthUrl(
		strconv.FormatInt(c.Chat().ID, 10),
		controller.authRedirectUrl,
	)

	if err != nil {
		_, _ = fmt.Fprintf(controller.out, "failed to get Auth url: %v\n", err)
		return err
	}

	output := bytes.Buffer{}
	err = controller.templates.ExecuteTemplate(&output, "WelcomeAnonymous.html", authUrl)
	if err == nil {
		err = c.Send(output.String())
	}
	return err
}

func (controller *TelegramController) welcomeAuthorizedAction(event *events.UserAuthorizedEvent) {
	student := controller.userRepository.GetStudent(event.ClientUserId)
	output := bytes.Buffer{}
	err := controller.templates.ExecuteTemplate(
		&output, "WelcomeAuthorizedAction.html",
		&UserAuthorizedTemplateData{student.GetTemplateData()},
	)

	if err == nil {
		_, err = controller.bot.Send(makeChatId(event.ClientUserId), output.String(), controller.markups.authorizedUserReplyMarkup)
	}

	if err != nil {
		_, _ = fmt.Fprintf(controller.out, "Failed to send welcome auth: %v", err)
	}
}

func (controller *TelegramController) logoutFinishedAction(event *events.UserAuthorizedEvent) {
	output := bytes.Buffer{}
	err := controller.templates.ExecuteTemplate(&output, "LogoutFinishedAction.html", nil)

	if err == nil {
		_, err = controller.bot.Send(makeChatId(event.ClientUserId), output.String(), controller.markups.logoutUserReplyMarkup)
	}

	if err != nil {
		_, _ = fmt.Fprintf(controller.out, "Failed to send logout finished: %v", err)
	}
}

func (controller *TelegramController) disciplinesListAction(c tele.Context) error {
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

		output := bytes.Buffer{}
		err = controller.templates.ExecuteTemplate(
			&output, "DisciplinesList.html",
			&DisciplinesListTemplateData{
				student.GetTemplateData(),
				disciplines,
			},
		)

		if err == nil {
			err = c.Send(output.String(), &telebot.SendOptions{
				DisableWebPagePreview: true,
			}, replyMarkup)
		}
	}

	return err
}

func (controller *TelegramController) disciplineScoresAction(c tele.Context) error {
	output := bytes.Buffer{}
	student := getStudent(c)
	disciplineId, _ := strconv.Atoi(c.Callback().Data)

	discipline, err := controller.scoreClient.GetStudentDiscipline(student.Id, disciplineId)

	if err == nil {
		err = controller.templates.ExecuteTemplate(
			&output, "DisciplineScore.html",
			&DisciplinesScoresTemplateData{
				student.GetTemplateData(),
				discipline,
			},
		)
	}

	if err == nil {
		err = c.Send(output.String(), controller.markups.disciplineScoreReplyMarkup)
	}

	return err
}

func (controller *TelegramController) scoreChangedAction(event *events.ScoreChangedEvent) {
	chatIds := controller.userRepository.GetClientUserIds(event.StudentId)

	for _, chatId := range chatIds {
		_, _ = controller.bot.Send(makeChatId(chatId), "S")
	}
}
