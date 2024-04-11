import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { MatToolbar } from '@angular/material/toolbar';
import { MatIcon, MatIconRegistry } from '@angular/material/icon';
import { MatButton, MatIconButton } from '@angular/material/button';
import { OidcSecurityService } from 'angular-auth-oidc-client';
import { MatMenu, MatMenuItem, MatMenuTrigger } from '@angular/material/menu';
import { MatList, MatListItem } from '@angular/material/list';
import {
  combineLatest,
  map,
  Observable,
  of,
  Subscription,
  switchMap,
  throwError,
} from 'rxjs';
import { UserInfoService } from '../services/user-info.service';
import { UserInfo } from '../interfaces/user';
import { DomSanitizer } from '@angular/platform-browser';
import { MatDialog } from '@angular/material/dialog';
import { SessionDialogComponent } from '../components/session-dialog/session-dialog.component';
import { SessionInfoService } from '../services/session-info.service';
import { SessionInfo } from '../interfaces/session';
import { MatCard, MatCardContent, MatCardHeader } from '@angular/material/card';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [
    MatToolbar,
    MatIcon,
    MatIconButton,
    MatMenuTrigger,
    MatMenu,
    MatList,
    MatMenuItem,
    MatListItem,
    MatButton,
    MatCard,
    MatCardContent,
    MatCardHeader,
  ],
  templateUrl: 'home.page.html',
  styleUrl: 'home.page.css',
})
export default class HomePage implements OnInit, OnDestroy {
  public isAuth: boolean = false;
  public userInfo: UserInfo | undefined;
  public isLoading: boolean = true;
  public sessionInfo: SessionInfo[] | undefined;

  private isAuth$: Observable<boolean> = of(false);
  private userInfo$: Observable<UserInfo> = of({ ID: 0, Email: '' });
  private isLoading$: Observable<boolean> = of(true);
  private sessionInfo$: Observable<SessionInfo[]> = of([]);
  private accessToken$: Observable<string> = of('');

  private authSubscription: Subscription = new Subscription();
  private userInfoSubscription: Subscription = new Subscription();
  private isLoadingSubscription: Subscription = new Subscription();
  private sessionInfoSubscription: Subscription = new Subscription();

  private readonly oidcSecurityService = inject(OidcSecurityService);
  private readonly userInfoService = inject(UserInfoService);
  private readonly sessionInfoService = inject(SessionInfoService);
  private readonly matIconRegistry = inject(MatIconRegistry);
  private readonly domSanitizer = inject(DomSanitizer);
  private readonly dialog = inject(MatDialog);

  constructor() {
    this.matIconRegistry.addSvgIcon(
      `aya_logo`,
      this.domSanitizer.bypassSecurityTrustResourceUrl('/aya.svg')
    );
  }

  ngOnInit(): void {
    let loginAttempt$ = this.oidcSecurityService.checkAuth();

    this.userInfo$ = loginAttempt$.pipe(
      switchMap((loginResponse) => {
        if (loginResponse.userData && loginResponse.accessToken) {
          let email = loginResponse.userData['email'];
          return this.userInfoService.getUserInfo$(
            loginResponse.accessToken,
            email
          );
        } else {
          return throwError(
            () => new Error('No user data yet. Please log in!')
          );
        }
      }),
      map(({ data, err }) => {
        if (err) {
          throw new Error(err);
        } else if (data) {
          return data;
        } else {
          throw new Error('No data found');
        }
      })
    );

    this.isAuth$ = loginAttempt$.pipe(
      map((loginAttempt) => loginAttempt.isAuthenticated)
    );

    this.accessToken$ = loginAttempt$.pipe(
      map((loginAttempt) => loginAttempt.accessToken)
    );

    this.sessionInfo$ = combineLatest([this.userInfo$, this.accessToken$]).pipe(
      switchMap(([userInfo, accessToken]) => {
        console.log('get all session?');
        return this.sessionInfoService.getAllSessions$(
          accessToken,
          userInfo.ID
        );
      }),
      map(({ data, err }) => {
        if (err) {
          throw new Error(err);
        } else if (data) {
          return data;
        } else {
          throw new Error('No data found');
        }
      })
    );

    this.isLoading$ = combineLatest([
      loginAttempt$,
      this.userInfo$,
      this.sessionInfo$,
    ]).pipe(
      map(([loginAttempt, userInfo, sessionsInfo]) => {
        return false;
      })
    );

    this.sessionInfoSubscription.add(
      this.sessionInfo$.subscribe({
        next: (sessionsInfo) => {
          this.sessionInfo = sessionsInfo;
        },
        error: (err) => {
          console.error('Cannot retrieve session info from backend');
          console.error(err);
        },
      })
    );

    this.userInfoSubscription.add(
      this.userInfo$.subscribe({
        next: (userInfo) => {
          this.userInfo = userInfo;
        },
        error: (err) => {
          console.error('Cannot retrieve user info from backend');
          console.error(err);
        },
      })
    );

    this.authSubscription.add(
      this.isAuth$.subscribe({
        next: (isAuth) => {
          this.isAuth = isAuth;
        },
        error: (err) => {
          console.error(err);
        },
      })
    );

    this.isLoadingSubscription.add(
      this.isLoading$.subscribe({
        next: (isLoading) => {
          this.isLoading = isLoading;
        },
        error: (err) => {
          console.log(err);
        },
      })
    );
  }

  login() {
    this.oidcSecurityService.authorize();
  }

  logout() {
    this.oidcSecurityService
      .logoff()
      .subscribe((result) => console.log(result));
  }

  ngOnDestroy(): void {
    this.authSubscription.unsubscribe();
    this.userInfoSubscription.unsubscribe();
    this.isLoadingSubscription.unsubscribe();
    this.sessionInfoSubscription.unsubscribe();
  }

  openDialog() {
    const dialogRef = this.dialog.open(SessionDialogComponent);
    dialogRef.afterClosed().subscribe((sessionInfoDialog) => {
      if (sessionInfoDialog) {
        combineLatest([this.accessToken$, this.userInfo$]).subscribe({
          next: ([accessToken, userInfo]) => {
            return this.sessionInfoService
              .newSession$(accessToken, userInfo.ID, sessionInfoDialog)
              .subscribe({
                next: (result) => {
                  console.log(result);
                },
                error: (err) => {
                  console.error(err);
                },
              });
          },
          error: (err) => {
            console.error(err);
          },
        });
      }
    });
  }

  protected readonly open = open;
}
