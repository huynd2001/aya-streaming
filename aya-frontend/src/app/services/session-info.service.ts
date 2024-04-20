import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SessionDialogInfo, SessionInfo } from '../interfaces/session';
import { catchError, Observable, of } from 'rxjs';

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
        catchError(
          (err: any): Observable<{ data?: SessionInfo[]; err?: string }> => {
            console.error(err);
            return of({
              data: undefined,
              err: String(err),
            });
          },
        ),
      );
  }

  newSession$(
    accessToken: string,
    userId: number,
    sessionDialogInfo: SessionDialogInfo,
  ) {
    return this.http.post<{ data?: SessionInfo; err?: string }>(
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
    );
  }

  updateSession$(
    accessToken: string,
    userId: number,
    sessionDialogInfo: SessionDialogInfo,
  ) {
    return this.http.put<{ data?: SessionInfo; err?: string }>(
      sessionInfoUrl,
      {
        user_id: userId,
        is_on: true,
        id: sessionDialogInfo.id,
        resources: JSON.stringify(sessionDialogInfo.resources),
      },
      {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      },
    );
  }
}
