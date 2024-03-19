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
	Id         string
	Author     Author
	Content    []MessagePart
	Attachment []string
}

type Emoji struct {
	Id  string
	Alt string
}

type Format struct {
	Color string
}

type MessagePart struct {
	CleanContent string
	Emoji        Emoji
	Format       Format
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
