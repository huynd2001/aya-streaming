package service

type ChatEmitter interface {
	UpdateEmitter() chan MessageUpdate
}
