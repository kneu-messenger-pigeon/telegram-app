package main

import (
	"github.com/kneu-messenger-pigeon/client-framework/models"
	tele "gopkg.in/telebot.v3"
	"strconv"
)

func getStudent(c tele.Context) *models.Student {
	return c.Get(contextStudentKey).(*models.Student)
}

func makeChatId(chatId string) tele.ChatID {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 0)
	return tele.ChatID(chatIdInt)
}
