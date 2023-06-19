package main

import (
	"github.com/kneu-messenger-pigeon/client-framework/models"
	tele "gopkg.in/telebot.v3"
	"strconv"
	"strings"
)

func getStudent(c tele.Context) *models.Student {
	return c.Get(contextStudentKey).(*models.Student)
}

func makeChatId(chatId string) tele.ChatID {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 0)
	return tele.ChatID(chatIdInt)
}

func makeInt64(input string) int64 {
	output, _ := strconv.ParseInt(input, 10, 0)
	return output
}

func escapeMarkDown(markdownStr string) string {
	// not "[", "]", "(", ")", - it is link
	escapeChar := []string{"#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range escapeChar {
		if strings.Contains(markdownStr, char) {
			markdownStr = strings.ReplaceAll(markdownStr, char, "\\"+char)
		}
	}
	escapeWithSpaceChar := []string{"("}
	for _, char := range escapeWithSpaceChar {
		if strings.Contains(markdownStr, " "+char) {
			markdownStr = strings.ReplaceAll(markdownStr, " "+char, " \\"+char)
		}
	}
	return markdownStr
}
