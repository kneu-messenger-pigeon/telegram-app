package main

import "github.com/VictoriaMetrics/metrics"

var (
	OnErrorCount         = metrics.NewCounter(`error_count{type="onError"}`)
	OnUpdateErrorCount   = metrics.NewCounter(`error_count{type="onUpdate"}`)
	RateLimitErrorCount  = metrics.NewCounter(`error_count{type="rateLimit"}`)
	TooManyRequestsCount = metrics.NewCounter(`error_count{type="tooManyRequests"}`)

	DisciplinesListActionRequestTotal  = metrics.NewCounter(`request_total{type="DisciplinesListAction"}`)
	DisciplineScoresActionRequestTotal = metrics.NewCounter(`request_total{type="DisciplineScoresAction"}`)
)
