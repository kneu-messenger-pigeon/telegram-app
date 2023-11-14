package main

import (
	framework "github.com/kneu-messenger-pigeon/client-framework"
	"github.com/kneu-messenger-pigeon/client-framework/models"
	tele "gopkg.in/telebot.v3"
	"strconv"
)

const contextStudentKey = "student"

func getStudent(c tele.Context) *models.Student {
	student := c.Get(contextStudentKey)
	if student == nil {
		return nil
	}
	return student.(*models.Student)
}

func authMiddleware(userRepository framework.UserRepositoryInterface) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			student := userRepository.GetStudent(strconv.FormatInt(c.Sender().ID, 10))
			if student != nil {
				c.Set(contextStudentKey, student)
			}

			return next(c)
		}
	}
}

func onlyAuthorizedMiddleware(anonymousHandler tele.HandlerFunc) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if getStudent(c) != nil {
				return next(c)
			}

			return anonymousHandler(c)
		}
	}
}

func onlyPrivateChatMiddleware() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Chat().Type == tele.ChatPrivate {
				return next(c)
			}

			return nil
		}
	}
}
