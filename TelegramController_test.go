package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/h2non/gock"
	authorizerMocks "github.com/kneu-messenger-pigeon/authorizer-client/mocks"
	framework "github.com/kneu-messenger-pigeon/client-framework"
	"github.com/kneu-messenger-pigeon/client-framework/mocks"
	"github.com/kneu-messenger-pigeon/client-framework/models"
	"github.com/kneu-messenger-pigeon/events"
	scoreApi "github.com/kneu-messenger-pigeon/score-api"
	"github.com/kneu-messenger-pigeon/score-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tele "gopkg.in/telebot.v3"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testTelegramURL = "http://telegram.test"
const testTelegramToken = "_TEST-token_"
const testTelegramUserId = int64(1238989)
const testTelegramUserIdString = "1238989"
const testTelegramSendMessageId = 99123456
const testTelegramIncomingMessageId = 123

var testPref = tele.Settings{
	Token:   testTelegramToken,
	URL:     testTelegramURL,
	Offline: true,
	Poller: &tele.LongPoller{
		Timeout: time.Minute,
	},
	ParseMode:   tele.ModeMarkdown,
	Synchronous: true,
}

func getTestSampleMessage() tele.Message {
	return tele.Message{
		ID:   testTelegramIncomingMessageId,
		Text: "",
		Sender: &tele.User{
			ID: testTelegramUserId,
		},
		Chat: &tele.Chat{
			ID:   testTelegramUserId,
			Type: tele.ChatPrivate,
		},
	}
}

var sampleStudent = &models.Student{
	Id:         uint32(999),
	LastName:   "Потапенко",
	FirstName:  "Андрій",
	MiddleName: "Петрович",
	Gender:     models.Student_MALE,
}

var testMessageText = "test-message ! 0101"

var sendMessageSuccessResponse = map[string]interface{}{
	"ok": true,
	"result": map[string]interface{}{
		"message_id": testTelegramSendMessageId,
	},
}

var lastTelegramErr error

func NewGock() *gock.Request {
	return gock.New(testTelegramURL + "/" + "bot" + testTelegramToken)
}

func CreateTelegramController(t *testing.T) (telegramController *TelegramController) {
	testPref.OnError = func(err error, c tele.Context) {
		lastTelegramErr = err
	}
	bot, _ := tele.NewBot(testPref)

	defer gock.Off()
	NewGock().Times(0)

	messageCompose := mocks.NewMessageComposerInterface(t)
	messageCompose.On("SetPostFilter", mock.AnythingOfType("func(string) string")).Once().Return()

	telegramController = &TelegramController{
		out:               &bytes.Buffer{},
		debugLogger:       &framework.DebugLogger{},
		bot:               bot,
		composer:          messageCompose,
		userRepository:    mocks.NewUserRepositoryInterface(t),
		userLogoutHandler: mocks.NewUserLogoutHandlerInterface(t),
		authorizerClient:  authorizerMocks.NewClientInterface(t),
		scoreClient:       score.NewMockClientInterface(t),
	}
	telegramController.Init()

	assert.True(t, gock.IsDone())

	t.Cleanup(func() {
		assert.NoError(t, lastTelegramErr)
	})

	return
}

func GetEndClearLastTelegramError() error {
	err := lastTelegramErr
	lastTelegramErr = nil
	return err
}

func TestTelegramController_NotPrivateChat(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		telegramController := CreateTelegramController(t)

		defer gock.Off()
		NewGock().Times(0)

		message := getTestSampleMessage()
		message.Text = startCommand
		message.Chat.Type = tele.ChatGroup

		telegramController.bot.ProcessUpdate(tele.Update{Message: &message})
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_Init(t *testing.T) {
	telegramController := CreateTelegramController(t)

	markups := &telegramController.markups

	assert.NotEmpty(t, markups.listButton)
	assert.NotEmpty(t, markups.listButton.Unique)

	assert.NotEmpty(t, markups.disciplineButton)
	assert.NotEmpty(t, markups.disciplineButton.Unique)

	assert.NotEmpty(t, markups.logoutUserReplyMarkup)
	assert.NotEmpty(t, markups.logoutUserReplyMarkup.ReplyKeyboard)
	assert.True(t, strings.HasPrefix(markups.logoutUserReplyMarkup.ReplyKeyboard[0][0].Text, startCommand))

	assert.NotEmpty(t, markups.authorizedUserReplyMarkup)
	assert.True(t, strings.HasPrefix(markups.authorizedUserReplyMarkup.ReplyKeyboard[0][0].Text, listCommand))

	assert.NotEmpty(t, markups.disciplineScoreReplyMarkup)
	assert.True(t, strings.HasPrefix(markups.authorizedUserReplyMarkup.ReplyKeyboard[0][0].Text, listCommand))
}

func TestTelegramController_ResetAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		telegramController := CreateTelegramController(t)

		userLogoutHandler := telegramController.userLogoutHandler.(*mocks.UserLogoutHandlerInterface)
		userLogoutHandler.On("Handle", testTelegramUserIdString).Return(nil).Once()

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		defer gock.Off()
		NewGock().Times(0)

		message := getTestSampleMessage()
		message.Text = resetCommand

		telegramController.bot.ProcessUpdate(tele.Update{Message: &message})
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_WelcomeAnonymousAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testAuthUrl := "http://auth.kneu.test/oauth"

		telegramController := CreateTelegramController(t)

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		authorizerClient := telegramController.authorizerClient.(*authorizerMocks.ClientInterface)
		authorizerClient.On("GetAuthUrl", testTelegramUserIdString, "https://t.me/?start").Return(testAuthUrl, nil)

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeWelcomeAnonymousMessage", testAuthUrl).Return(nil, testMessageText)

		sendMessageRequest := map[string]interface{}{
			"chat_id":         testTelegramUserIdString,
			"parse_mode":      "Markdown",
			"protect_content": "true",
			"text":            testMessageText,
		}

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(sendMessageRequest).
			Reply(200).JSON(sendMessageSuccessResponse)

		message := getTestSampleMessage()
		message.Text = startCommand

		telegramController.bot.ProcessUpdate(tele.Update{Message: &message})

		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})

	t.Run("authUrlError", func(t *testing.T) {
		telegramController := CreateTelegramController(t)
		expectedError := errors.New("expected error")

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		authorizerClient := telegramController.authorizerClient.(*authorizerMocks.ClientInterface)
		authorizerClient.On("GetAuthUrl", testTelegramUserIdString, "https://t.me/?start").Return("", expectedError)

		defer gock.Off()
		NewGock().Times(0)

		out := &bytes.Buffer{}
		telegramController.out = out

		message := getTestSampleMessage()
		message.Text = startCommand

		telegramController.bot.ProcessUpdate(tele.Update{Message: &message})

		assert.Equal(t, expectedError, GetEndClearLastTelegramError())
		assert.Contains(t, out.String(), expectedError.Error())
		assert.True(t, gock.IsDone())
	})

}

func TestTelegramController_WelcomeAuthorizedAction(t *testing.T) {
	messageData := models.UserAuthorizedMessageData{
		StudentMessageData: models.StudentMessageData{
			NamePrefix: "Пане",
			Name:       "",
		},
	}

	telegramController := CreateTelegramController(t)

	expectedJson := map[string]interface{}{
		"chat_id":      testTelegramUserIdString,
		"parse_mode":   "Markdown",
		"reply_markup": toJson(telegramController.markups.authorizedUserReplyMarkup),
		"text":         testMessageText,
	}

	event := &events.UserAuthorizedEvent{
		Client:       "test",
		ClientUserId: testTelegramUserIdString,
		StudentId:    1234,
		LastName:     "",
		FirstName:    "",
		MiddleName:   "",
		Gender:       0,
	}

	t.Run("success", func(t *testing.T) {
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeWelcomeAuthorizedMessage", messageData).Return(nil, testMessageText)

		telegramController.composer = messageCompose
		telegramController.userRepository = userRepository

		defer gock.Off()
		NewGock().Times(1).
			Post("/sendMessage").JSON(expectedJson).
			Reply(200).JSON(sendMessageSuccessResponse)

		err := telegramController.WelcomeAuthorizedAction(event)

		assert.NoError(t, err)
		assert.True(t, gock.IsDone())
	})

	t.Run("telegramError", func(t *testing.T) {
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeWelcomeAuthorizedMessage", messageData).Return(nil, testMessageText)

		telegramController.composer = messageCompose
		telegramController.userRepository = userRepository
		out := telegramController.out.(*bytes.Buffer)

		errorText := "example-test-error"

		defer gock.Off()
		NewGock().Times(1).
			Post("/sendMessage").JSON(expectedJson).
			Reply(400).JSON(map[string]interface{}{
			"ok":          false,
			"error_code":  400,
			"description": errorText,
		})

		err := telegramController.WelcomeAuthorizedAction(event)
		assert.Error(t, err)
		assert.Contains(t, out.String(), errorText)
		assert.Contains(t, out.String(), testMessageText)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_LogoutFinishedAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		telegramController := CreateTelegramController(t)

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeLogoutFinishedMessage").Return(nil, testMessageText)

		expectedJson := map[string]interface{}{
			"chat_id":      testTelegramUserIdString,
			"parse_mode":   "Markdown",
			"reply_markup": toJson(telegramController.markups.logoutUserReplyMarkup),
			"text":         testMessageText,
		}

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(expectedJson).
			Reply(200).JSON(sendMessageSuccessResponse)

		event := &events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: testTelegramUserIdString,
			StudentId:    0,
			LastName:     "",
			FirstName:    "",
			MiddleName:   "",
			Gender:       0,
		}

		err := telegramController.LogoutFinishedAction(event)

		assert.NoError(t, err)
		assert.True(t, gock.IsDone())
	})

	t.Run("error", func(t *testing.T) {
		expectedError := errors.New("expected error")

		telegramController := CreateTelegramController(t)

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeLogoutFinishedMessage").Return(expectedError, "")

		defer gock.Off()
		NewGock().Times(0)

		event := &events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: testTelegramUserIdString,
			StudentId:    0,
			LastName:     "",
			FirstName:    "",
			MiddleName:   "",
			Gender:       0,
		}

		err := telegramController.LogoutFinishedAction(event)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)

		assert.True(t, gock.IsDone())
	})

	t.Run("error_send", func(t *testing.T) {
		errorText := "Bad Request: can't parse entities"
		expectedError := errors.New("telegram: " + errorText + " (400)")

		telegramController := CreateTelegramController(t)
		out := telegramController.out.(*bytes.Buffer)

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeLogoutFinishedMessage").Return(nil, testMessageText)

		expectedJson := map[string]interface{}{
			"chat_id":      testTelegramUserIdString,
			"parse_mode":   "Markdown",
			"reply_markup": toJson(telegramController.markups.logoutUserReplyMarkup),
			"text":         testMessageText,
		}

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(expectedJson).
			Reply(400).
			JSON(map[string]interface{}{
				"ok":          false,
				"error_code":  400,
				"description": errorText,
			})

		event := &events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: testTelegramUserIdString,
			StudentId:    0,
			LastName:     "",
			FirstName:    "",
			MiddleName:   "",
			Gender:       0,
		}

		err := telegramController.LogoutFinishedAction(event)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Contains(t, out.String(), errorText)
		assert.Contains(t, out.String(), testMessageText)

		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_DisciplinesListAction(t *testing.T) {
	disciplines := scoreApi.DisciplineScoreResults{
		{
			Discipline: scoreApi.Discipline{
				Id:   100,
				Name: "Капітал!",
			},
			ScoreRating: scoreApi.ScoreRating{
				Total:         17,
				StudentsCount: 25,
				Rating:        8,
				MinTotal:      10,
				MaxTotal:      20,
			},
		},
		{
			Discipline: scoreApi.Discipline{
				Id:   110,
				Name: "Гроші та лихварство",
			},
			ScoreRating: scoreApi.ScoreRating{
				Total:         12,
				StudentsCount: 25,
				Rating:        12,
				MinTotal:      7,
				MaxTotal:      17,
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		telegramController := CreateTelegramController(t)

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := telegramController.scoreClient.(*score.MockClientInterface)
		scoreClient.On("GetStudentDisciplines", sampleStudent.Id).Return(disciplines, nil)

		messageData := models.DisciplinesListMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Disciplines:        disciplines,
		}

		replyMarkup := &tele.ReplyMarkup{
			OneTimeKeyboard: true,
			InlineKeyboard:  make([][]tele.InlineButton, len(disciplines)),
		}
		for i, discipline := range disciplines {
			replyMarkup.InlineKeyboard[i] = []tele.InlineButton{
				{
					Unique: telegramController.markups.disciplineButton.Unique,
					Data:   strconv.Itoa(discipline.Discipline.Id),
					Text:   discipline.Discipline.Name,
				},
			}
		}
		ProcessReplyMarkup(replyMarkup)

		expectedMessageSend := map[string]interface{}{
			"chat_id":      testTelegramUserIdString,
			"parse_mode":   "Markdown",
			"reply_markup": toJson(replyMarkup),
			"text":         testMessageText,
		}

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeDisciplinesListMessage", messageData).Return(nil, testMessageText)

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(expectedMessageSend).
			Reply(200).JSON(sendMessageSuccessResponse)

		message := getTestSampleMessage()
		message.Text = listCommand

		telegramController.bot.ProcessUpdate(tele.Update{Message: &message})

		assert.True(t, gock.IsDone())
	})

	t.Run("error", func(t *testing.T) {
		telegramController := CreateTelegramController(t)
		expectedError := errors.New("expected error")

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := telegramController.scoreClient.(*score.MockClientInterface)
		scoreClient.On("GetStudentDisciplines", sampleStudent.Id).Return(disciplines, nil)

		messageData := models.DisciplinesListMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Disciplines:        disciplines,
		}

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeDisciplinesListMessage", messageData).Return(expectedError, "")

		defer gock.Off()
		NewGock().Post("/sendMessage").Times(0)

		message := getTestSampleMessage()
		message.Text = listCommand

		telegramController.bot.ProcessUpdate(tele.Update{Message: &message})

		assert.Equal(t, expectedError, GetEndClearLastTelegramError())
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_DisciplineScoresAction(t *testing.T) {
	disciplineId := 199

	discipline := scoreApi.DisciplineScoreResult{
		Discipline: scoreApi.Discipline{
			Id:   disciplineId,
			Name: "Капітал!",
		},
		ScoreRating: scoreApi.ScoreRating{
			Total:         17,
			StudentsCount: 25,
			Rating:        8,
			MinTotal:      10,
			MaxTotal:      20,
		},
		Scores: []scoreApi.Score{
			{
				Lesson: scoreApi.Lesson{
					Id:   245,
					Date: time.Date(2023, time.Month(2), 12, 0, 0, 0, 0, time.Local),
					Type: scoreApi.LessonType{
						Id:        5,
						ShortName: "МК",
						LongName:  "Модульний контроль.",
					},
				},
				FirstScore: floatPointer(4.5),
				IsAbsent:   true,
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		telegramController := CreateTelegramController(t)

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := telegramController.scoreClient.(*score.MockClientInterface)
		scoreClient.On("GetStudentDiscipline", sampleStudent.Id, disciplineId).Return(discipline, nil)

		messageData := models.DisciplinesScoresMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Discipline:         discipline,
		}
		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeDisciplineScoresMessage", messageData).Return(nil, testMessageText)

		replyMarkup := &tele.ReplyMarkup{
			OneTimeKeyboard: true,
			InlineKeyboard: [][]tele.InlineButton{
				{
					*telegramController.markups.listButton,
				},
			},
		}
		ProcessReplyMarkup(replyMarkup)

		expectedJson := map[string]interface{}{
			"chat_id":      testTelegramUserIdString,
			"parse_mode":   "Markdown",
			"reply_markup": toJson(replyMarkup),
			"text":         testMessageText,
		}

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(expectedJson).
			Reply(200).JSON(sendMessageSuccessResponse)

		cbData := fmt.Sprintf(`%s|%d`, telegramController.markups.disciplineButton.CallbackUnique(), disciplineId)

		message := getTestSampleMessage()
		message.Text = ""

		telegramController.bot.ProcessUpdate(tele.Update{
			Message: &message,
			Callback: &tele.Callback{
				Data:   cbData,
				Sender: message.Sender,
			},
		})

		assert.True(t, gock.IsDone())
	})

	t.Run("error", func(t *testing.T) {
		telegramController := CreateTelegramController(t)
		expectedError := errors.New("expected error")

		userRepository := telegramController.userRepository.(*mocks.UserRepositoryInterface)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := telegramController.scoreClient.(*score.MockClientInterface)
		scoreClient.On("GetStudentDiscipline", sampleStudent.Id, disciplineId).Return(scoreApi.DisciplineScoreResult{}, expectedError)

		out := &bytes.Buffer{}
		telegramController.out = out

		expectedJson := map[string]interface{}{
			"chat_id":      testTelegramUserIdString,
			"message_id":   strconv.Itoa(testTelegramIncomingMessageId),
			"reply_markup": "{}",
		}

		defer gock.Off()
		NewGock().Post("/sendMessage").Times(0)

		NewGock().
			Post("/editMessageReplyMarkup").JSON(expectedJson).Times(1).
			Reply(400).
			JSON(map[string]interface{}{
				"ok":          false,
				"error_code":  400,
				"description": "Bad Request: message is not modified",
			})

		button := telegramController.markups.disciplineButton.With(strconv.Itoa(disciplineId))
		ProcessInlineButton(button)

		message := getTestSampleMessage()
		message.Text = ""

		telegramController.bot.ProcessUpdate(tele.Update{
			Message: &message,
			Callback: &tele.Callback{
				Data:   button.Data,
				Sender: message.Sender,
			},
		})

		assert.Equal(t, expectedError, GetEndClearLastTelegramError())
		assert.True(t, gock.IsDone())
		assert.Contains(t, out.String(), `Failed to remove reply markup: telegram: message is not modified (400)`)
	})
}

func TestTelegramController_ScoreChangedAction(t *testing.T) {
	telegramController := CreateTelegramController(t)

	discipline := scoreApi.Discipline{
		Id:   12,
		Name: "Капітал!",
	}

	// input values
	disciplineScore := &scoreApi.DisciplineScore{
		Discipline: discipline,
		Score: scoreApi.Score{
			Lesson: scoreApi.Lesson{
				Id:   150,
				Date: time.Date(2023, time.Month(2), 12, 0, 0, 0, 0, time.Local),
				Type: scoreApi.LessonType{
					Id:        5,
					ShortName: "МК",
					LongName:  "Модульний контроль.",
				},
			},
			FirstScore: floatPointer(2.5),
		},
	}

	previousScore := &scoreApi.Score{}

	//
	disciplineButton := telegramController.markups.disciplineButton.With(strconv.Itoa(disciplineScore.Discipline.Id))
	disciplineButton.Text = disciplineScore.Discipline.Name

	replyMarkup := &tele.ReplyMarkup{
		OneTimeKeyboard: true,
		InlineKeyboard:  [][]tele.InlineButton{{*disciplineButton}},
	}
	ProcessReplyMarkup(replyMarkup)
	replyMarkupJson := toJson(replyMarkup)

	expectedSendMessage := map[string]interface{}{
		"chat_id":      testTelegramUserIdString,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkupJson,
		"text":         testMessageText,
	}
	//

	messageData := models.ScoreChangedMessageData{
		Discipline: disciplineScore.Discipline,
		Score:      disciplineScore.Score,
		Previous:   *previousScore,
	}

	t.Run("send_new_message", func(t *testing.T) {
		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeScoreChanged", messageData).Return(nil, testMessageText)

		defer gock.Off()
		NewGock().
			Times(1).
			Post("/sendMessage").
			JSON(expectedSendMessage).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		actualErr, actualMessageId := telegramController.ScoreChangedAction(
			testTelegramUserIdString, "", disciplineScore, previousScore,
		)
		assert.NoError(t, actualErr)
		assert.True(t, gock.IsDone())
		assert.Equal(t, strconv.Itoa(testTelegramSendMessageId), actualMessageId)
	})

	t.Run("edit_previous_message", func(t *testing.T) {
		var previousChatMessageId = "6655443322"

		thisCaseExpectedMessageSend := map[string]interface{}{
			"chat_id":      testTelegramUserIdString,
			"message_id":   previousChatMessageId,
			"parse_mode":   "Markdown",
			"reply_markup": replyMarkupJson,
			"text":         testMessageText,
		}

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeScoreChanged", messageData).Return(nil, testMessageText)

		defer gock.Off()
		NewGock().Times(1).Post("/editMessageText").JSON(thisCaseExpectedMessageSend).
			Reply(200).JSON(sendMessageSuccessResponse)

		actualErr, actualMessageId := telegramController.ScoreChangedAction(
			testTelegramUserIdString, previousChatMessageId, disciplineScore, previousScore,
		)
		assert.NoError(t, actualErr)
		assert.True(t, gock.IsDone())
		assert.Equal(t, strconv.Itoa(testTelegramSendMessageId), actualMessageId)
	})

	t.Run("delete_previous_message", func(t *testing.T) {
		t.Run("previous_message_exist", func(t *testing.T) {
			var previousChatMessageId = "6655443322"

			expectedJson := map[string]interface{}{
				"chat_id":    testTelegramUserIdString,
				"message_id": previousChatMessageId,
			}

			deleteMessageSuccessResponse := map[string]interface{}{
				"ok":     true,
				"result": true,
			}

			thisCasePreviousScore := &scoreApi.Score{
				FirstScore: floatPointer(2.5),
			}

			thisCaseMessageData := models.ScoreChangedMessageData{
				Discipline: disciplineScore.Discipline,
				Score:      disciplineScore.Score,
				Previous:   *thisCasePreviousScore,
			}

			messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
			messageCompose.On("ComposeScoreChanged", thisCaseMessageData).Return(nil, testMessageText)

			defer gock.Off()
			NewGock().Times(1).Post("/deleteMessage").JSON(expectedJson).
				Reply(200).JSON(deleteMessageSuccessResponse)

			actualErr, actualMessageId := telegramController.ScoreChangedAction(
				testTelegramUserIdString, previousChatMessageId, disciplineScore, thisCasePreviousScore,
			)
			assert.NoError(t, actualErr)
			assert.True(t, gock.IsDone())
			assert.Empty(t, actualMessageId)
		})

		t.Run("previous_message_no_exist", func(t *testing.T) {
			var previousChatMessageId = ""

			thisCasePreviousScore := &scoreApi.Score{
				FirstScore: floatPointer(2.5),
			}

			thisCaseMessageData := models.ScoreChangedMessageData{
				Discipline: disciplineScore.Discipline,
				Score:      disciplineScore.Score,
				Previous:   *thisCasePreviousScore,
			}

			messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
			messageCompose.On("ComposeScoreChanged", thisCaseMessageData).Return(nil, testMessageText)

			defer gock.Off()
			NewGock().Post("/sendMessage").Times(0)

			actualErr, actualMessageId := telegramController.ScoreChangedAction(
				testTelegramUserIdString, previousChatMessageId, disciplineScore, thisCasePreviousScore,
			)
			assert.NoError(t, actualErr)
			assert.True(t, gock.IsDone())
			assert.Empty(t, actualMessageId)
		})

	})

	t.Run("error", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}

		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeScoreChanged", messageData).Return(nil, testMessageText)

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(expectedSendMessage).Reply(400)

		actualErr, actualMessageId := telegramController.ScoreChangedAction(
			testTelegramUserIdString, "", disciplineScore, previousScore,
		)
		assert.Error(t, actualErr)
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
		assert.Empty(t, actualMessageId)
	})

	t.Run("error-bot-blocked-by-user", func(t *testing.T) {
		messageCompose := telegramController.composer.(*mocks.MessageComposerInterface)
		messageCompose.On("ComposeScoreChanged", messageData).Return(nil, testMessageText)

		userLogoutHandler := telegramController.userLogoutHandler.(*mocks.UserLogoutHandlerInterface)
		userLogoutHandler.On("Handle", testTelegramUserIdString).Return(nil).Once()

		defer gock.Off()
		NewGock().Times(1).Post("/sendMessage").JSON(expectedSendMessage).
			Reply(400).JSON(map[string]interface{}{
			"ok":          false,
			"error_code":  403,
			"description": "Forbidden: bot was blocked by the user",
		})

		actualErr, actualMessageId := telegramController.ScoreChangedAction(
			testTelegramUserIdString, "", disciplineScore, previousScore,
		)
		assert.NoError(t, actualErr)
		assert.True(t, gock.IsDone())
		assert.Empty(t, actualMessageId)
	})
}

func floatPointer(value float32) *float32 {
	return &value
}

func toJson(v interface{}) string {
	jsonData, _ := json.Marshal(v)
	return string(jsonData)
}

func ProcessReplyMarkup(markup *tele.ReplyMarkup) {
	result := tele.ResultBase{
		ParseMode:   tele.ModeMarkdown,
		ReplyMarkup: markup,
	}
	result.Process(nil)
}

func ProcessInlineButton(button *tele.InlineButton) {
	replyMarkup := &tele.ReplyMarkup{
		OneTimeKeyboard: true,
		InlineKeyboard: [][]tele.InlineButton{
			{*button},
		},
	}
	ProcessReplyMarkup(replyMarkup)

	*button = replyMarkup.InlineKeyboard[0][0]
}

func PrintBodyMatcher(req *http.Request, ereq *gock.Request) (bool, error) {
	fmt.Println(req.URL.String())
	bodyBytes, _ := io.ReadAll(req.Body)
	fmt.Println(string(bodyBytes))

	return true, nil
}
