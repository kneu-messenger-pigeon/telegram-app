package main

import "github.com/VictoriaMetrics/metrics"

var (
	onErrorCount       = metrics.NewCounter(`error_count{type="onError"}`)
	onUpdateErrorCount = metrics.NewCounter(`error_count{type="onUpdate"}`)

	DisciplinesListActionRequestTotal  = metrics.NewCounter(`request_total{type="DisciplinesListAction"}`)
	DisciplineScoresActionRequestTotal = metrics.NewCounter(`request_total{type="DisciplineScoresAction"}`)
)
