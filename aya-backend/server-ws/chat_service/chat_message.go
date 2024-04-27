package chat_service

import (
	"encoding/json"
	"fmt"
	"time"
)

type Source int

const (
	Discord Source = iota
	Twitch
	Youtube
	TestSource
)

var (
	sourceToStr = map[int]string{
		0: "discord",
		1: "twitch",
		2: "youtube",
		3: "test_source",
	}

	strToSource = map[string]int{
		"discord":     0,
		"twitch":      1,
		"youtube":     2,
		"test_source": 3,
	}
)

func ParseSource(s string) (Source, error) {
	value, ok := strToSource[s]
	if !ok {
		return Source(-1), fmt.Errorf(`cannot detect "%s", not a valid source`, s)
	}
	return Source(value), nil
}

func (s Source) String() string {
	return sourceToStr[int(s)]
}

func (s Source) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Source) UnmarshalJSON(data []byte) error {
	var source string
	var err error
	if err = json.Unmarshal(data, &source); err != nil {
		return err
	}
	if *s, err = ParseSource(source); err != nil {
		return err
	}
	return nil
}

type Update int

const (
	New Update = iota
	Delete
	Edit
)

var (
	updateToStr = map[int]string{
		0: "new",
		1: "delete",
		2: "edit",
	}

	strToUpdate = map[string]int{
		"new":    0,
		"delete": 1,
		"edit":   2,
	}
)

func ParseUpdate(s string) (Update, error) {
	value, ok := strToUpdate[s]
	if !ok {
		return Update(-1), fmt.Errorf(`cannot detect "%s", not a valid source`, s)
	}
	return Update(value), nil
}

func (s Update) String() string {
	return updateToStr[int(s)]
}

func (s Update) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Attachment string

func (s *Update) UnmarshalJSON(data []byte) error {
	var update string
	var err error
	if err = json.Unmarshal(data, &update); err != nil {
		return err
	}
	if *s, err = ParseUpdate(update); err != nil {
		return err
	}
	return nil
}

type Message struct {
	Source       Source        `json:"source"`
	Id           string        `json:"id"`
	Author       Author        `json:"author"`
	MessageParts []MessagePart `json:"messageParts"`
	Attachments  []Attachment  `json:"attachments"`
}

type Emoji struct {
	Id  string `json:"id,omitempty"`
	Alt string `json:"alt,omitempty"`
}

type Format struct {
	Color string `json:"color,omitempty"`
}

type MessagePart struct {
	Content string  `json:"content"`
	Emoji   *Emoji  `json:"emoji,omitempty"`
	Format  *Format `json:"format,omitempty"`
}

type Author struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	IsBot    bool   `json:"isBot"`
	Color    string `json:"color"`
}

type MessageUpdate struct {
	UpdateTime  time.Time `json:"updateTime"`
	Update      Update    `json:"update"`
	Message     Message   `json:"message"`
	ExtraFields any       `json:"-"`
}
