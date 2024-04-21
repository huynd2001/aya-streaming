package hubs

type SessionResourceHub interface {
	GetSessionId(resourceInfo any) []string
	RemoveSession(sessionId string)
	AddSession(sessionId string)
}
