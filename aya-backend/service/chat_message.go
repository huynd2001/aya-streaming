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
	Id         string        `json:"id"`
	Author     Author        `json:"author"`
	Content    []MessagePart `json:"content"`
	Attachment []string      `json:"attachment"`
}

type Emoji struct {
	Id  string `json:"id"`
	Alt string `json:"alt"`
}

type Format struct {
	Color string `json:"color"`
}

type MessagePart struct {
	CleanContent string `json:"cleanContent"`
	Emoji        Emoji  `json:"emoji"`
	Format       Format `json:"format"`
}

type Author struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	IsBot    bool   `json:"isBot"`
	Color    string `json:"color"`
}

type MessageUpdate struct {
	Source  Source  `json:"source"`
	Update  Update  `json:"update"`
	Message Message `json:"message"`
}
