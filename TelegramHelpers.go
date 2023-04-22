package main

import (
	tele "gopkg.in/telebot.v3"
	"strconv"
)

func getStudent(c tele.Context) *Student {
	return c.Get(contextStudentKey).(*Student)
}

func makeChatId(chatId string) tele.ChatID {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 0)
	return tele.ChatID(chatIdInt)
}
