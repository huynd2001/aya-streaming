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
import { SessionInfo } from '../interfaces/session';
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

  constructor(private _snackBar: MatSnackBar) {
    this.matIconRegistry.addSvgIcon(
      `aya_logo`,
      this.domSanitizer.bypassSecurityTrustResourceUrl('/aya.svg'),
    );
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
          this.sessionInfo = sessionsInfo;
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
              return this.sessionInfoService.newSession$(
                accessToken,
                userInfo.ID,
                sessionInfoDialog,
              );
            }),
          )
          .subscribe({
            next: (value) => {
              console.log(value);
            },
            error: (err) => {
              console.error(err);
            },
          });
      }
    });
  }

  openEditDialog(id: number) {
    if (!this.sessionInfo) {
      return;
    }
    if (id < 0 || id >= this.sessionInfo.length) {
      return;
    }

    const dialogRef = this.dialog.open(SessionDialogComponent, {
      data: {
        id: this.sessionInfo[id].ID,
        resources: JSON.parse(this.sessionInfo[id].Resources),
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
    if (!this.sessionInfo) {
      return;
    }
    if (id < 0 || id >= this.sessionInfo.length) {
      return;
    }
    const dialogRef = this.dialog.open(YesNoDialogComponent);
    let sessionId = this.sessionInfo[id].ID;
    dialogRef.afterClosed().subscribe((userCollect) => {
      if (userCollect === true) {
        combineLatest([this.accessToken$, this.userInfo$])
          .pipe(
            switchMap(([accessToken, userInfo]) => {
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

  protected readonly open = open;
}
