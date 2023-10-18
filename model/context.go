package model

type ContextKey int

const (
	LOGGER ContextKey = iota
	REQUEST_INFO
)
