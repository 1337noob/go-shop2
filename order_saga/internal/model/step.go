package model

import (
	"shop/pkg/command"
	"shop/pkg/event"
)

type StepStatus string

const (
	StepStatusInit      StepStatus = "init"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
)

type Step struct {
	Command                command.Type `json:"command"`
	CommandStatus          StepStatus   `json:"command_status"`
	CommandSuccessEvent    event.Type   `json:"command_success_event"`
	CommandFailEvent       event.Type   `json:"command_fail_event"`
	Compensate             command.Type `json:"compensate"`
	CompensateStatus       StepStatus   `json:"compensate_status"`
	CompensateSuccessEvent event.Type   `json:"compensate_success_event"`
	CompensateFailEvent    event.Type   `json:"compensate_fail_event"`
	CommandTopic           string       `json:"command_topic"`
}
