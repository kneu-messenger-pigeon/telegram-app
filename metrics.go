package main

import "github.com/VictoriaMetrics/metrics"

var (
	OnErrorCount         = metrics.NewCounter(`error_count{type="onError"}`)
	OnUpdateErrorCount   = metrics.NewCounter(`error_count{type="onUpdate"}`)
	TooManyRequestsCount = metrics.NewCounter(`error_count{type="tooManyRequests"}`)

	DisciplinesListActionRequestTotal  = metrics.NewCounter(`request_total{type="DisciplinesListAction"}`)
	DisciplineScoresActionRequestTotal = metrics.NewCounter(`request_total{type="DisciplineScoresAction"}`)
)
