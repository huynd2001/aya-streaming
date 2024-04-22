import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SessionDialogInfo, SessionInfo } from '../interfaces/session';
import { catchError, map, Observable, of } from 'rxjs';

const sessionInfoUrl = `/api/session/`;

@Injectable({
  providedIn: 'root',
})
export class SessionInfoService {
  constructor(private readonly http: HttpClient) {}

  getAllSessions$(accessToken: string, userId: number) {
    return this.http
      .get<{ data?: SessionInfo[]; err?: string }>(sessionInfoUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
        params: {
          user_id: userId,
        },
      })
      .pipe(
        map((result) => {
          if (!result) {
            throw Error('result cannot be parsed');
          }
          if (result?.err) {
            throw result.err;
          } else {
            return {
              data: result.data,
              err: undefined,
            };
          }
        }),
      );
  }

  newSession$(
    accessToken: string,
    userId: number,
    sessionDialogInfo: SessionDialogInfo,
  ) {
    return this.http
      .post<{ data?: SessionInfo; err?: string }>(
        sessionInfoUrl,
        {
          user_id: userId,
          resources: JSON.stringify(sessionDialogInfo.resources),
        },
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        },
      )
      .pipe(
        map((result) => {
          if (!result) {
            throw Error('result cannot be parsed');
          }
          if (result?.err) {
            throw result.err;
          } else {
            return {
              data: result.data,
              err: undefined,
            };
          }
        }),
      );
  }

  updateSession$(
    accessToken: string,
    userId: number,
    sessionDialogInfo: SessionDialogInfo,
    isOn?: boolean,
  ) {
    return this.http
      .put<{ data?: SessionInfo; err?: string }>(
        sessionInfoUrl,
        {
          user_id: userId,
          is_on: isOn,
          id: sessionDialogInfo.id,
          resources: JSON.stringify(sessionDialogInfo.resources),
        },
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        },
      )
      .pipe(
        map((result) => {
          if (!result) {
            throw Error('result cannot be parsed');
          }
          if (result?.err) {
            throw result.err;
          } else {
            return {
              data: result.data,
              err: undefined,
            };
          }
        }),
      );
  }

  deleteSession$(accessToken: string, userId: number, sessionId: number) {
    return this.http
      .delete<{ data?: SessionInfo; err?: string }>(sessionInfoUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
        params: {
          user_id: userId,
          id: sessionId,
        },
      })
      .pipe(
        map(
          (
            result,
          ): { data: SessionInfo | undefined; err: string | undefined } => {
            if (!result) {
              throw Error('result cannot be parsed');
            }
            if (result.err) {
              throw Error(result.err);
            } else {
              return { data: result.data, err: undefined };
            }
          },
        ),
      );
  }
}
