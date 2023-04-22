package main

import (
	scoreApi "github.com/kneu-messenger-pigeon/score-api"
)

type StudentBaseTemplateData struct {
	NamePrefix string
	Name       string
}

type UserAuthorizedTemplateData struct {
	StudentBaseTemplateData
}

type DisciplinesListTemplateData struct {
	StudentBaseTemplateData
	Disciplines scoreApi.DisciplineScoreResults
}

type DisciplinesScoresTemplateData struct {
	StudentBaseTemplateData
	Discipline scoreApi.DisciplineScoreResult
}
