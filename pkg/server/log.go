package server

type Logger interface {
	Info(a ...interface{})
	Error(a ...interface{})
}
