package main

import (
	framework "github.com/kneu-messenger-pigeon/client-framework"
	tele "gopkg.in/telebot.v3"
	"strconv"
)

func getStudent(c tele.Context) *framework.Student {
	return c.Get(contextStudentKey).(*framework.Student)
}

func makeChatId(chatId string) tele.ChatID {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 0)
	return tele.ChatID(chatIdInt)
}
