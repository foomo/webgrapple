package clientnpm

type Logger interface {
	Info(a ...interface{})
	Error(a ...interface{})
}
