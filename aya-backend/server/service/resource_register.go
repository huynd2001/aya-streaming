package service

type ResourceRegister interface {
	Register(resourceInfo any)
	Deregister(resourceInfo any)
}
