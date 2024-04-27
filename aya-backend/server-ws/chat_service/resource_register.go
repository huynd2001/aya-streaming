package chat_service

type ResourceRegister interface {
	Register(subscriber string, resourceInfo any)
	Deregister(subscriber string, resourceInfo any)
}
