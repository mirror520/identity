package model

import (
	"encoding/json"
	"errors"
	"time"
)

type ResultStatus string

const (
	SUCCESS ResultStatus = "success"
	FAILURE ResultStatus = "failure"
)

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

type Result struct {
	Status ResultStatus
	Msg    string
	Data   any
	Raw    json.RawMessage
	Time   time.Time
}

func (r *Result) Bytes() ([]byte, error) {
	return json.Marshal(&r)
}

func (r *Result) Error() error {
	if r.Status == SUCCESS {
		return nil
	}

	return errors.New(r.Msg)
}

func (r *Result) MarshalJSON() ([]byte, error) {
	output := struct {
		Status ResultStatus    `json:"status"`
		Msg    string          `json:"msg"`
		Data   json.RawMessage `json:"data"`
		Time   time.Time       `json:"time"`
	}{
		Status: r.Status,
		Msg:    r.Msg,
		Data:   nil,
		Time:   r.Time,
	}

	if r.Data != nil {
		bs, err := json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}

		output.Data = bs
	}

	return json.Marshal(&output)
}

func (r *Result) UnmarshalJSON(data []byte) error {
	var input struct {
		Status ResultStatus    `json:"status"`
		Msg    string          `json:"msg"`
		Data   json.RawMessage `json:"data"`
		Time   time.Time       `json:"time"`
	}

	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	r.Status = input.Status
	r.Msg = input.Msg
	r.Raw = input.Data
	r.Time = input.Time

	return nil
}
