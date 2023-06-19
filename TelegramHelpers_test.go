package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_EscapeMarkDown(t *testing.T) {
	t.Run("usual", func(t *testing.T) {
		input := "Задачі прикладного системного аналізу: 50\nрейтинг #1/20."
		expected := "Задачі прикладного системного аналізу: 50\nрейтинг \\#1/20\\."

		assert.Equal(t, expected, escapeMarkDown(input))
	})

	t.Run("brackets-with-leading-space", func(t *testing.T) {
		input := "Бали: 3 (було 1)"
		expected := "Бали: 3 \\(було 1)"

		assert.Equal(t, expected, escapeMarkDown(input))
	})

	t.Run("brackets-without-leading-space", func(t *testing.T) {
		input := "Бали: 3(було 1)"
		expected := "Бали: 3(було 1)"

		assert.Equal(t, expected, escapeMarkDown(input))
	})

}
