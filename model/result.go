package model

import "time"

type ResultStatus string

const (
	SUCCESS ResultStatus = "success"
	FAILURE ResultStatus = "failure"
)

type Result struct {
	Status ResultStatus `json:"status"`
	Msg    string       `json:"msg"`
	Data   any          `json:"data"`
	Time   time.Time    `json:"time"`
}

func SuccessResult(msg string) *Result {
	return &Result{
		Status: SUCCESS,
		Msg:    msg,
		Data:   nil,
		Time:   time.Now(),
	}
}

func FailureResult(err error) *Result {
	return &Result{
		Status: FAILURE,
		Msg:    err.Error(),
		Time:   time.Now(),
	}
}
