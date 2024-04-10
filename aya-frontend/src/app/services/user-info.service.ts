import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { User } from '../interfaces/user';
import { catchError, Observable, of, switchMap } from 'rxjs';

const userInfoUrl = `/api/user/`;

@Injectable({
  providedIn: 'root',
})
export class UserInfoService {
  constructor(private http: HttpClient) {}

  getUserInfo$(accessToken: string, email: string) {
    return this.http
      .get<{ data?: User; err?: string }>(userInfoUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
        params: {
          email: email,
        },
      })
      .pipe(
        catchError((err: any): Observable<{ data?: User; err?: string }> => {
          console.error(err);
          return of({
            data: undefined,
            err: String(err),
          });
        }),
        switchMap(({ data, err }) => {
          if (!data && err === 'Cannot find the profile') {
            return this.newUser$(accessToken, email);
          } else {
            return of({
              data: data,
              err: err,
            });
          }
        })
      );
  }

  newUser$(accessToken: string, email: string) {
    return this.http
      .post<{ data?: User; err?: string }>(
        userInfoUrl,
        {
          email: email,
        },
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        }
      )
      .pipe(
        catchError((err: any) => {
          console.error(err);
          return of({
            data: undefined,
            err: String(err),
          });
        })
      );
  }
}
