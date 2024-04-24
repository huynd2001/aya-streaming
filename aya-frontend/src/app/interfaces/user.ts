import { SessionInfo } from './session';

export interface UserInfo {
  ID: number;
  Email: string;
  Sessions: SessionInfo[];
}

export interface UserReq {
  id?: number;
  email: string;
}
