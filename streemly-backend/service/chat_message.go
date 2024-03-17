package service

type Source int

const (
	Discord Source = iota
	Twitch
	Youtube
	TestSource
)

type Update int

const (
	New Update = iota
	Delete
	Edit
)

type Message struct {
	Source     Source
	Id         string
	Author     Author
	Content    []MessagePart
	Attachment []string
}

type MessagePart struct {
	CleanContent string

	Emoji struct {
		Id  string
		Alt string
	}

	Format struct {
		Color string
	}
}

type Author struct {
	Username string
	IsAdmin  bool
	IsBot    bool
	Color    string
}

type MessageUpdate struct {
	Source  Source
	Update  Update
	Message Message
}
