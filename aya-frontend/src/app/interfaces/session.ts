export interface SessionDialogInfo {
  id?: number;
  resources: ResourceInfo[];
}

export interface SessionInfo {
  ID: number;
  Resources: string;
  IsOn: boolean;
  IsDelete: boolean;
  UserID: number;
}

export interface ResourceInfo {
  resourceType: string;
  resourceInfo: {
    discordChannelId?: string;
    discordGuildId?: string;
    youtubeChannelId?: string;
  };
}
