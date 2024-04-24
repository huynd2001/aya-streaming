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
  shareReplay,
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
import {
  SessionDialogInfo,
  SessionInfo,
  DisplaySessionInfo,
} from '../interfaces/session';
import {
  MatCard,
  MatCardActions,
  MatCardContent,
  MatCardHeader,
} from '@angular/material/card';
import { MatDivider } from '@angular/material/divider';
import { YesNoDialogComponent } from '../components/yes-no-dialog/yes-no-dialog.component';
import { MatSnackBar } from '@angular/material/snack-bar';
import { SessionInfoDisplayComponent } from '../components/session-info-display/session-info-display.component';
import {
  MatSlideToggle,
  MatSlideToggleChange,
} from '@angular/material/slide-toggle';
import { injectRouter } from '@analogjs/router';
import { Clipboard } from '@angular/cdk/clipboard';
import { environment } from '../../environments/environment';

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
    MatCardActions,
    MatDivider,
    SessionInfoDisplayComponent,
    MatSlideToggle,
  ],
  templateUrl: 'home.page.html',
  styleUrl: 'home.page.css',
})
export default class HomePage implements OnInit, OnDestroy {
  public isAuth: boolean = false;
  public userInfo: UserInfo | undefined;
  public isLoading: boolean = true;
  public displaySessionInfo: DisplaySessionInfo[] | undefined;

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
  private readonly router = injectRouter();

  constructor(
    private _snackBar: MatSnackBar,
    private clipboard: Clipboard,
  ) {
    this.matIconRegistry.addSvgIcon(
      `aya_logo`,
      this.domSanitizer.bypassSecurityTrustResourceUrl('/aya.svg'),
    );
  }

  copyURLToClipboard(streamUUID: string) {
    this.clipboard.copy(`${environment.homepageUrl}/stream/${streamUUID}`);
    this._snackBar.open(`Session stream link copied!`, 'Dismiss', {
      duration: 3000,
    });
  }

  ngOnInit(): void {
    let loginAttempt$ = this.oidcSecurityService
      .checkAuth()
      .pipe(shareReplay(1));

    this.userInfo$ = loginAttempt$.pipe(
      switchMap((loginResponse) => {
        if (loginResponse.userData && loginResponse.accessToken) {
          let email = loginResponse.userData['email'];
          return this.userInfoService.getUserInfo$(
            loginResponse.accessToken,
            email,
          );
        } else {
          return throwError(
            () => new Error('No user data yet. Please log in!'),
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
      }),
      shareReplay(1),
    );

    this.isAuth$ = loginAttempt$.pipe(
      map((loginAttempt) => loginAttempt.isAuthenticated),
      shareReplay(1),
    );

    this.accessToken$ = loginAttempt$.pipe(
      map((loginAttempt) => loginAttempt.accessToken),
      shareReplay(1),
    );

    this.sessionInfo$ = combineLatest([this.userInfo$, this.accessToken$]).pipe(
      switchMap(([userInfo, accessToken]) => {
        return this.sessionInfoService.getAllSessions$(
          accessToken,
          userInfo.ID,
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
      }),
      shareReplay(1),
    );

    this.isLoading$ = combineLatest([
      loginAttempt$,
      this.userInfo$,
      this.sessionInfo$,
    ]).pipe(
      map(([loginAttempt, userInfo, sessionsInfo]) => {
        return false;
      }),
      shareReplay(1),
    );

    this.sessionInfoSubscription.add(
      this.sessionInfo$.subscribe({
        next: (sessionsInfo) => {
          this.displaySessionInfo = sessionsInfo.map(
            (sessionInfo): DisplaySessionInfo => {
              return {
                session_info: sessionInfo,
                should_hidden: false,
              };
            },
          );
        },
        error: (err) => {
          console.error('Cannot retrieve session info from backend');
          console.error(err);
        },
      }),
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
      }),
    );

    this.authSubscription.add(
      this.isAuth$.subscribe({
        next: (isAuth) => {
          this.isAuth = isAuth;
        },
        error: (err) => {
          console.error(err);
        },
      }),
    );

    this.isLoadingSubscription.add(
      this.isLoading$.subscribe({
        next: (isLoading) => {
          this.isLoading = isLoading;
        },
        error: (err) => {
          console.log(err);
        },
      }),
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
        combineLatest([this.accessToken$, this.userInfo$])
          .pipe(
            switchMap(([accessToken, userInfo]) => {
              this._snackBar.open('Submitting...', 'Dismiss');
              return this.sessionInfoService.newSession$(
                accessToken,
                userInfo.ID,
                sessionInfoDialog,
              );
            }),
            map((data) => {
              if (data.err) {
                throw data.err;
              } else {
                return data.data;
              }
            }),
          )
          .subscribe({
            next: (value) => {
              const sessionId = value?.ID;
              this._snackBar.open(
                `Create Session ${sessionId} Success!`,
                'Dismiss',
                {
                  duration: 3000,
                },
              );
            },
            error: (err) => {
              console.error(err);
            },
          });
      }
    });
  }

  openEditDialog(id: number) {
    if (!this.displaySessionInfo) {
      return;
    }
    if (id < 0 || id >= this.displaySessionInfo.length) {
      return;
    }

    const dialogRef = this.dialog.open(SessionDialogComponent, {
      data: {
        id: this.displaySessionInfo[id].session_info.ID,
        resources: JSON.parse(
          this.displaySessionInfo[id].session_info.Resources,
        ),
      },
    });

    dialogRef.afterClosed().subscribe((sessionInfoDialog) => {
      if (sessionInfoDialog) {
        const sessionId = sessionInfoDialog.id;
        combineLatest([this.accessToken$, this.userInfo$])
          .pipe(
            switchMap(([accessToken, userInfo]) => {
              this._snackBar.open('Loading...', 'Dismiss');
              return this.sessionInfoService.updateSession$(
                accessToken,
                userInfo.ID,
                sessionInfoDialog,
              );
            }),
          )
          .subscribe({
            next: (value) => {
              this._snackBar.open(
                `Update Session ${sessionId} Success!`,
                'Dismiss',
                {
                  duration: 3000,
                },
              );
              if (this.displaySessionInfo && value.data) {
                this.displaySessionInfo[id] = {
                  should_hidden: false,
                  session_info: value.data,
                };
              }
            },
            error: (err) => {
              console.error(err);
              this._snackBar.open(
                `Update Session ${sessionId} Failed!`,
                'Dismiss',
                {
                  duration: 3000,
                },
              );
            },
          });
      }
    });
  }

  openDeleteDialog(id: number) {
    if (!this.displaySessionInfo) {
      return;
    }
    if (id < 0 || id >= this.displaySessionInfo.length) {
      return;
    }
    const dialogRef = this.dialog.open(YesNoDialogComponent);
    let sessionId = this.displaySessionInfo[id].session_info.ID;
    dialogRef.afterClosed().subscribe((userCollect) => {
      if (userCollect === true) {
        combineLatest([this.accessToken$, this.userInfo$])
          .pipe(
            switchMap(([accessToken, userInfo]) => {
              this._snackBar.open('Loading...', 'Dismiss');
              return this.sessionInfoService.deleteSession$(
                accessToken,
                userInfo.ID,
                sessionId,
              );
            }),
          )
          .subscribe({
            next: (value) => {
              this._snackBar.open(
                `Delete Session ${sessionId} Success!`,
                'Dismiss',
                {
                  duration: 3000,
                },
              );
              if (this.displaySessionInfo && value.data) {
                this.displaySessionInfo[id].should_hidden = true;
              }
            },
            error: (err) => {
              console.error(err);
              this._snackBar.open(
                `Delete Session ${sessionId} Failed!`,
                'Dismiss',
                {
                  duration: 3000,
                },
              );
            },
          });
      }
    });
  }

  switchSession(id: number, event: MatSlideToggleChange) {
    if (!this.displaySessionInfo) {
      return;
    }
    if (id < 0 || id >= this.displaySessionInfo.length) {
      return;
    }
    const sessionId = this.displaySessionInfo[id].session_info.ID;
    const newSessionInfo: SessionDialogInfo = {
      id: this.displaySessionInfo[id].session_info.ID,
      resources: JSON.parse(this.displaySessionInfo[id].session_info.Resources),
    };
    combineLatest([this.accessToken$, this.userInfo$])
      .pipe(
        switchMap(([accessToken, userInfo]) => {
          this._snackBar.open('Loading...', 'Dismiss');
          return this.sessionInfoService.updateSession$(
            accessToken,
            userInfo.ID,
            newSessionInfo,
            event.checked,
          );
        }),
      )
      .subscribe({
        next: (value) => {
          this._snackBar.open(
            `Update Session ${sessionId} Success!`,
            'Dismiss',
            {
              duration: 3000,
            },
          );
          if (this.displaySessionInfo && value.data) {
            this.displaySessionInfo[id] = {
              should_hidden: false,
              session_info: value.data,
            };
          }
        },
        error: (err) => {
          console.error(err);
          this._snackBar.open(
            `Update Session ${sessionId} Failed!`,
            'Dismiss',
            {
              duration: 3000,
            },
          );
        },
      });
  }

  protected readonly open = open;
}
