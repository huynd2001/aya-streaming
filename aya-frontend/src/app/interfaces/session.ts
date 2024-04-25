import { WritableSignal } from '@angular/core';

export interface SessionDialogInfo {
  id?: number;
  resources: ResourceInfo[];
}

export interface SessionInfo {
  ID: number;
  UUID: string;
  Resources: string;
  IsOn: boolean;
  IsDelete: boolean;
  UserID: number;
}

export interface DisplaySessionInfo {
  should_hidden: boolean;
  session_info: WritableSignal<SessionInfo>;
}

export interface ResourceInfo {
  resourceType: string;
  resourceInfo: {
    discordChannelId?: string;
    discordGuildId?: string;
    youtubeChannelId?: string;
  };
}
