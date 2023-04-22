package main

import (
	"fmt"
	scoreApi "github.com/kneu-messenger-pigeon/score-api"
	"html/template"
	"strconv"
	"time"
)

var TemplateFunctionMap = template.FuncMap{
	"incr": func(a int) int {
		return a + 1
	},

	"date": func(date time.Time) string {
		return date.Format("02.01.2006")
	},

	"renderScore": func(score scoreApi.Score) string {
		if score.FirstScore != 0 && score.SecondScore != 0 {
			return fmt.Sprintf(
				"%s та %s",
				strconv.FormatFloat(float64(score.FirstScore), 'f', -1, 32),
				strconv.FormatFloat(float64(score.SecondScore), 'f', -1, 32),
			)
		}

		if score.FirstScore != 0 {
			return strconv.FormatFloat(float64(score.FirstScore), 'f', -1, 32)
		}

		if score.SecondScore != 0 {
			return strconv.FormatFloat(float64(score.SecondScore), 'f', -1, 32)
		}

		if score.IsAbsent {
			return "пропуск"
		}

		return "0"
	},
}
