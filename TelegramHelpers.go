package main

import (
	"github.com/kneu-messenger-pigeon/client-framework/models"
	tele "gopkg.in/telebot.v3"
	"regexp"
	"strconv"
	"strings"
)

func getStudent(c tele.Context) *models.Student {
	student := c.Get(contextStudentKey)
	if student == nil {
		return nil
	}

	return student.(*models.Student)
}

func makeChatId(chatId string) tele.ChatID {
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 0)
	return tele.ChatID(chatIdInt)
}

func makeInt64(input string) int64 {
	output, _ := strconv.ParseInt(input, 10, 0)
	return output
}

/** @see https://regex101.com/r/5zFBzu/1 */
var unEscapeMarkDownLinks = regexp.MustCompile(`(?m)\\\[([^\[\]]*)\\\]\\\(([^\)]*)\\\)`)
var unEscapeMarkDownLinksSubstitution = "[$1]($2)"

func escapeMarkDown(markdownStr string) string {
	// do not escape special chars to keep format: * bold; _ italic; ~strikethrough; | - spoiler;
	escapeChar := []string{"[", "]", "(", ")", ">", "#", "+", "-", "=", "{", "}", ".", "!", "`"}
	for _, char := range escapeChar {
		if strings.Contains(markdownStr, char) {
			markdownStr = strings.ReplaceAll(markdownStr, char, "\\"+char)
		}
	}

	// drop escaping from inline links
	markdownStr = unEscapeMarkDownLinks.ReplaceAllString(markdownStr, unEscapeMarkDownLinksSubstitution)

	return markdownStr
}
