package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/h2non/gock"
	authorizerMocks "github.com/kneu-messenger-pigeon/authorizer-client/mocks"
	"github.com/kneu-messenger-pigeon/client-framework/mocks"
	"github.com/kneu-messenger-pigeon/client-framework/models"
	"github.com/kneu-messenger-pigeon/events"
	scoreApi "github.com/kneu-messenger-pigeon/score-api"
	scoreMocks "github.com/kneu-messenger-pigeon/score-client/mocks"
	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v3"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

const testTelegramURL = "http://telegram.test"
const testTelegramToken = "_TEST-token_"
const testTelegramUserId = int64(1238989)
const testTelegramUserIdString = "1238989"

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

var sendMessageRequest = `{"chat_id":"1238989","parse_mode":"Markdown","text":"test-message ! 0101"}`

var sendMessageSuccessResponse = `{"ok":true,"message":{"id":123456}}`

func TestTelegramController_NotPrivateChat(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).Times(0)

		telegramController := &TelegramController{
			out: &bytes.Buffer{},
			bot: bot,
		}
		telegramController.Init()

		message := getTestSampleMessage()
		message.Text = startCommand
		message.Chat.Type = tele.ChatGroup

		bot.ProcessUpdate(tele.Update{Message: &message})
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_ResetAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)

		userLogoutHandler := mocks.NewUserLogoutHandlerInterface(t)
		userLogoutHandler.On("Handle", testTelegramUserIdString).Return(nil).Once()

		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).Times(0)

		telegramController := &TelegramController{
			out:               &bytes.Buffer{},
			bot:               bot,
			userRepository:    userRepository,
			userLogoutHandler: userLogoutHandler,
		}
		telegramController.Init()

		message := getTestSampleMessage()
		message.Text = resetCommand

		bot.ProcessUpdate(tele.Update{Message: &message})
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_WelcomeAnonymousAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testAuthUrl := "http://auth.kneu.test/oauth"

		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		authorizerClient := authorizerMocks.NewClientInterface(t)
		authorizerClient.On("GetAuthUrl", testTelegramUserIdString, "https://t.me/?start").Return(testAuthUrl, nil)

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeWelcomeAnonymousMessage", testAuthUrl).Return(nil, testMessageText)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(sendMessageRequest).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		telegramController := &TelegramController{
			out:              &bytes.Buffer{},
			bot:              bot,
			composer:         messageCompose,
			userRepository:   userRepository,
			authorizerClient: authorizerClient,
		}
		telegramController.Init()

		message := getTestSampleMessage()
		message.Text = "/start"

		bot.ProcessUpdate(tele.Update{Message: &message})

		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})

	t.Run("authUrlError", func(t *testing.T) {
		expectedError := errors.New("expected error")

		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		authorizerClient := authorizerMocks.NewClientInterface(t)
		authorizerClient.On("GetAuthUrl", testTelegramUserIdString, "https://t.me/?start").Return("", expectedError)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(0)

		out := &bytes.Buffer{}
		telegramController := &TelegramController{
			out:              out,
			bot:              bot,
			userRepository:   userRepository,
			authorizerClient: authorizerClient,
		}
		telegramController.Init()

		message := getTestSampleMessage()
		message.Text = "/start"

		bot.ProcessUpdate(tele.Update{Message: &message})

		assert.Error(t, lastTelegramErr)
		assert.Equal(t, expectedError, lastTelegramErr)
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

	replyMarkupJson := strings.Replace(
		regexp.QuoteMeta(`,"reply_markup":"{\"keyboard\":[[{\"text\":\"\"}]]}"`),
		`\\"text\\":\\"\\"`,
		`\\"text\\":\\"`+listCommand+`.*\\"`,
		1,
	)
	insertBefore := `,"text":`
	expectedMessage := strings.Replace(sendMessageRequest, insertBefore, replyMarkupJson+insertBefore, 1)

	t.Run("success", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeWelcomeAuthorizedMessage", messageData).Return(nil, testMessageText)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(expectedMessage).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		event := &events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: testTelegramUserIdString,
			StudentId:    1234,
			LastName:     "",
			FirstName:    "",
			MiddleName:   "",
			Gender:       0,
		}

		err := telegramController.WelcomeAuthorizedAction(event)

		assert.NoError(t, err)
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})

	t.Run("telegramError", func(t *testing.T) {
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(&models.Student{}).Once()

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeWelcomeAuthorizedMessage", messageData).Return(nil, testMessageText)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(expectedMessage).
			Reply(400)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		event := &events.UserAuthorizedEvent{
			Client:       "test",
			ClientUserId: testTelegramUserIdString,
			StudentId:    1234,
			LastName:     "",
			FirstName:    "",
			MiddleName:   "",
			Gender:       0,
		}

		err := telegramController.WelcomeAuthorizedAction(event)

		assert.Error(t, err)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_LogoutFinishedAction(t *testing.T) {
	replyMarkupJson := strings.Replace(
		regexp.QuoteMeta(`"reply_markup":"{\"keyboard\":[[{\"text\":\"\"}]]}",`),
		`\\"text\\":\\"\\"`,
		`\\"text\\":\\"`+startCommand+`.*\\"`,
		1,
	)
	insertBefore := `"text":`
	expectedMessage := strings.Replace(sendMessageRequest, insertBefore, replyMarkupJson+insertBefore, 1)

	t.Run("success", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeLogoutFinishedMessage").Return(nil, testMessageText)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(expectedMessage).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		telegramController := &TelegramController{
			out:      &bytes.Buffer{},
			bot:      bot,
			composer: messageCompose,
		}
		telegramController.Init()

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
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})

	t.Run("error", func(t *testing.T) {
		expectedError := errors.New("expected error")
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeLogoutFinishedMessage").Return(expectedError, "")

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(0)

		telegramController := &TelegramController{
			out:      &bytes.Buffer{},
			bot:      bot,
			composer: messageCompose,
		}
		telegramController.Init()

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

		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})
}

func TestTelegramController_DisciplinesListAction(t *testing.T) {
	expectedMessage := sendMessageRequest
	var insertBefore string

	replyMarkupJson := `"reply_markup":"{\\"inline_keyboard\\":(.*)}",`
	insertBefore = `"text":`
	expectedMessage = strings.Replace(expectedMessage, insertBefore, replyMarkupJson+insertBefore, 1)

	disableWebPagePreview := `"disable_web_page_preview":"true",`
	insertBefore = `"parse_mode":`
	expectedMessage = strings.Replace(expectedMessage, insertBefore, disableWebPagePreview+insertBefore, 1)

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
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := scoreMocks.NewClientInterface(t)
		scoreClient.On("GetStudentDisciplines", sampleStudent.Id).Return(disciplines, nil)

		messageData := models.DisciplinesListMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Disciplines:        disciplines,
		}

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeDisciplinesListMessage", messageData).Return(nil, testMessageText)

		bodyRegExp, _ := regexp.Compile(expectedMessage)
		actualMarkupJson := ""

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(expectedMessage).
			AddMatcher(func(request *http.Request, _ *gock.Request) (bool, error) {
				body, _ := io.ReadAll(request.Body)
				matches := bodyRegExp.FindStringSubmatch(string(body))
				if len(matches) >= 2 {
					actualMarkupJson = strings.Replace(matches[1], `\"`, `"`, -1)
					return true, nil
				}
				return false, nil
			}).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			scoreClient:    scoreClient,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		message := getTestSampleMessage()
		message.Text = listCommand

		bot.ProcessUpdate(tele.Update{Message: &message})

		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())

		var actualInlineButtons [][]tele.InlineButton
		err := json.Unmarshal([]byte(actualMarkupJson), &actualInlineButtons)
		assert.NoError(t, err)

		assert.Len(t, actualInlineButtons, len(disciplines))

		callbackPrefixBytes, _ := json.Marshal(telegramController.markups.disciplineButton.CallbackUnique())
		expectedCallback := strings.Trim(string(callbackPrefixBytes), `"`) + `|%d`

		var discipline scoreApi.Discipline
		for index, actualInlineButtonRow := range actualInlineButtons {
			assert.Len(t, actualInlineButtonRow, 1)

			discipline = disciplines[index].Discipline
			disciplineButton := actualInlineButtonRow[0]

			assert.Equal(t, fmt.Sprintf(expectedCallback, discipline.Id), disciplineButton.Data)
			assert.Equal(t, discipline.Name, disciplineButton.Text)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedError := errors.New("expected error")

		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := scoreMocks.NewClientInterface(t)
		scoreClient.On("GetStudentDisciplines", sampleStudent.Id).Return(disciplines, nil)

		messageData := models.DisciplinesListMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Disciplines:        disciplines,
		}

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeDisciplinesListMessage", messageData).Return(expectedError, "")

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Post("/sendMessage").
			Times(0)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			scoreClient:    scoreClient,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		message := getTestSampleMessage()
		message.Text = listCommand

		bot.ProcessUpdate(tele.Update{Message: &message})

		assert.Error(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
		assert.Equal(t, expectedError, lastTelegramErr)
	})
}

func TestTelegramController_DisciplineScoresAction(t *testing.T) {
	disciplineId := 199

	expectedMessage := sendMessageRequest
	var insertBefore string

	replyMarkupJson := `"reply_markup":"{\\"inline_keyboard\\":(.*)}",`
	insertBefore = `"text":`
	expectedMessage = strings.Replace(expectedMessage, insertBefore, replyMarkupJson+insertBefore, 1)

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
				FirstScore:  4.5,
				SecondScore: 0,
				IsAbsent:    true,
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := scoreMocks.NewClientInterface(t)
		scoreClient.On("GetStudentDiscipline", sampleStudent.Id, disciplineId).Return(discipline, nil)

		messageData := models.DisciplinesScoresMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Discipline:         discipline,
		}

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeDisciplineScoresMessage", messageData).Return(nil, testMessageText)

		bodyRegExp, _ := regexp.Compile(expectedMessage)
		actualMarkupJson := ""

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(expectedMessage).
			AddMatcher(func(request *http.Request, _ *gock.Request) (bool, error) {
				body, _ := io.ReadAll(request.Body)
				matches := bodyRegExp.FindStringSubmatch(string(body))
				if len(matches) >= 2 {
					actualMarkupJson = strings.Replace(matches[1], `\"`, `"`, -1)
					return true, nil
				}
				return false, nil
			}).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			scoreClient:    scoreClient,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		cbData := fmt.Sprintf(`%s|%d`, telegramController.markups.disciplineButton.CallbackUnique(), disciplineId)

		message := getTestSampleMessage()
		message.Text = ""

		bot.ProcessUpdate(tele.Update{
			Message: &message,
			Callback: &tele.Callback{
				Data:   cbData,
				Sender: message.Sender,
			},
		})

		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())

		var actualInlineButtons [][]tele.InlineButton
		err := json.Unmarshal([]byte(actualMarkupJson), &actualInlineButtons)
		assert.NoError(t, err)

		assert.Len(t, actualInlineButtons, 1)
		assert.Len(t, actualInlineButtons[0], 1)

		backButton := actualInlineButtons[0][0]

		assert.Equal(t, telegramController.markups.listButton.Unique, backButton.Unique)
		assert.Equal(t, telegramController.markups.listButton.Text, backButton.Text)

		expectedDataBytes, _ := json.Marshal(telegramController.markups.listButton.CallbackUnique())
		expectedData := strings.Trim(string(expectedDataBytes), `"`)
		assert.Equal(t, expectedData, backButton.Data)
	})

	t.Run("error", func(t *testing.T) {
		expectedError := errors.New("expected error")

		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)
		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetStudent", testTelegramUserIdString).Return(sampleStudent).Once()

		scoreClient := scoreMocks.NewClientInterface(t)
		scoreClient.On("GetStudentDiscipline", sampleStudent.Id, disciplineId).Return(discipline, nil)

		messageData := models.DisciplinesScoresMessageData{
			StudentMessageData: models.NewStudentMessageData(sampleStudent),
			Discipline:         discipline,
		}

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeDisciplineScoresMessage", messageData).Return(expectedError, "")

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Post("/sendMessage").
			Times(0)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			scoreClient:    scoreClient,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		cbData := fmt.Sprintf(`%s|%d`, telegramController.markups.disciplineButton.CallbackUnique(), disciplineId)

		message := getTestSampleMessage()
		message.Text = ""

		bot.ProcessUpdate(tele.Update{
			Message: &message,
			Callback: &tele.Callback{
				Data:   cbData,
				Sender: message.Sender,
			},
		})

		assert.Error(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
		assert.Equal(t, expectedError, lastTelegramErr)
	})
}

func TestTelegramController_ScoreChangedAction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var lastTelegramErr error
		testPref.OnError = func(err error, c tele.Context) {
			lastTelegramErr = err
		}
		bot, _ := tele.NewBot(testPref)

		scoreEvent := &events.ScoreChangedEvent{
			ScoreEvent: events.ScoreEvent{
				Id:           0,
				StudentId:    uint(sampleStudent.Id),
				LessonId:     0,
				LessonPart:   0,
				DisciplineId: 0,
				Year:         0,
				Semester:     0,
				Value:        0,
				IsAbsent:     false,
				IsDeleted:    false,
				UpdatedAt:    time.Time{},
				SyncedAt:     time.Time{},
			},
			Previous: struct {
				Value     float32
				IsAbsent  bool
				IsDeleted bool
			}{},
		}

		userRepository := mocks.NewUserRepositoryInterface(t)
		userRepository.On("GetClientUserIds", scoreEvent.StudentId).
			Return([]string{testTelegramUserIdString}).
			Once()

		messageCompose := mocks.NewMessageComposerInterface(t)
		messageCompose.On("ComposeScoreChanged").Return(nil, testMessageText)

		defer gock.Off()
		gock.New(testTelegramURL + "/" + "bot" + testTelegramToken).
			Times(1).
			Post("/sendMessage").
			JSON(sendMessageRequest).
			Reply(200).
			JSON(sendMessageSuccessResponse)

		telegramController := &TelegramController{
			out:            &bytes.Buffer{},
			bot:            bot,
			composer:       messageCompose,
			userRepository: userRepository,
		}
		telegramController.Init()

		err := telegramController.ScoreChangedAction(scoreEvent)
		assert.NoError(t, err)
		assert.NoError(t, lastTelegramErr)
		assert.True(t, gock.IsDone())
	})
}
