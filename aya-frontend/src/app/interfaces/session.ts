export interface SessionDialogInfo {
  id?: number;
  resources: {
    resourceType: string;
    resourceInfo: {
      discordChannelId?: string;
      discordGuildId?: string;
      youtubeChannelId?: string;
    };
  }[];
}

export interface SessionInfo {
  ID: number;
  Resources: string;
  IsOn: boolean;
  IsDelete: boolean;
  UserID: number;
}
