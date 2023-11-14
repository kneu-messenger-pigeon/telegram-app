package main

import (
	framework "github.com/kneu-messenger-pigeon/client-framework"
	tele "gopkg.in/telebot.v3"
	"strconv"
)

const contextStudentKey = "student"

func authMiddleware(userRepository framework.UserRepositoryInterface) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			c.Set(
				contextStudentKey,
				userRepository.GetStudent(
					strconv.FormatInt(c.Sender().ID, 10),
				),
			)

			return next(c)
		}
	}
}

func onlyAuthorizedMiddleware(anonymousHandler tele.HandlerFunc) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if getStudent(c).Id != 0 {
				return next(c)
			}

			return anonymousHandler(c)
		}
	}
}

func onlyPrivateChatMiddleware() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Chat() != nil && c.Chat().Type == tele.ChatPrivate {
				return next(c)
			}

			return nil
		}
	}
}
