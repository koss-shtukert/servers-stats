package common

type Notifier interface {
	SendMessage(m string)
}
