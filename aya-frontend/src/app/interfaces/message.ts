export interface Author {
  isBot: boolean;
  isAdmin: boolean;
  username: string;
  color: string;
}

export interface Format {
  color: string;
}

export interface Emoji {
  alt: string;
  id: string;
}

export interface MessagePart {
  format?: Format;
  emoji?: Emoji;
  content: string;
}

export interface MessageUpdate {
  update: 'new' | 'edit' | 'delete';
  message: Message;
}

export interface Message {
  source: 'discord' | 'youtube' | 'twitch' | 'test_source';
  id: string;
  author: Author;
  attachments: string[];
  messageParts: MessagePart[];
}

export interface DisplayMessage {
  message: Message;
  delete?: boolean;
  edit?: boolean;
  remove?: boolean;
  init?: boolean;
}
