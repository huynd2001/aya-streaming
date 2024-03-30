package service

type ChatEmitter interface {
	// UpdateEmitter returns a new channel that sends out the MessageUpdate
	UpdateEmitter() chan MessageUpdate
	// CloseEmitter everything and free up resources
	CloseEmitter() error
}
