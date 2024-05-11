import {
  Component,
  computed,
  inject,
  OnDestroy,
  OnInit,
  Signal,
  signal,
  WritableSignal,
} from '@angular/core';
import { MatToolbar } from '@angular/material/toolbar';
import { MatIcon, MatIconRegistry } from '@angular/material/icon';
import { MatButton, MatIconButton } from '@angular/material/button';
import { LoginResponse, OidcSecurityService } from 'angular-auth-oidc-client';
import { MatMenu, MatMenuItem, MatMenuTrigger } from '@angular/material/menu';
import { MatList, MatListItem } from '@angular/material/list';
import { map, Subscription } from 'rxjs';
import { UserInfoService } from '../services/user-info.service';
import { UserInfo } from '../interfaces/user';
import { DomSanitizer } from '@angular/platform-browser';
import { MatDialog } from '@angular/material/dialog';
import { SessionDialogComponent } from '../components/session-dialog/session-dialog.component';
import { SessionInfoService } from '../services/session-info.service';
import { DisplaySessionInfo, SessionDialogInfo } from '../interfaces/session';
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
  private loginResponse: WritableSignal<LoginResponse | undefined> =
    signal(undefined);
  public isAuth: Signal<boolean> = computed(
    () => this.loginResponse()?.isAuthenticated || false,
  );
  private accessToken: WritableSignal<string> = signal('');

  public userInfo: WritableSignal<UserInfo> = signal({
    ID: -1,
    Email: '',
    Sessions: [],
  });
  public displaySessionInfo: WritableSignal<DisplaySessionInfo[]> = signal([]);

  private userInfoSubscription: Subscription = new Subscription();
  private loginAttemptSubscription: Subscription = new Subscription();
  private accessTokenSubscription: Subscription = new Subscription();
  private updateDialogSubscription: Subscription = new Subscription();
  private deleteDialogSubscription: Subscription = new Subscription();
  private newDialogSubscription: Subscription = new Subscription();

  private readonly oidcSecurityService = inject(OidcSecurityService);
  private readonly userInfoService = inject(UserInfoService);
  private readonly sessionInfoService = inject(SessionInfoService);
  private readonly matIconRegistry = inject(MatIconRegistry);
  private readonly domSanitizer = inject(DomSanitizer);
  private readonly dialog = inject(MatDialog);

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
    this.loginAttemptSubscription.add(
      this.oidcSecurityService.checkAuth().subscribe((loginResponse) => {
        this.loginResponse.set(loginResponse);
        let email = loginResponse.userData['email'];
        this.userInfoSubscription.add(
          this.userInfoService
            .getUserInfo$(loginResponse.accessToken, email)
            .subscribe(({ data, err }) => {
              if (err) {
                console.error(err);
                return;
              } else if (!data) {
                console.error('cannot parse data');
              } else {
                this.userInfo.set(data);
                this.displaySessionInfo.set(
                  data.Sessions.map((sessionInfo): DisplaySessionInfo => {
                    return {
                      session_info: signal(sessionInfo),
                      should_hidden: false,
                    };
                  }),
                );
              }
            }),
        );
      }),
    );

    this.accessTokenSubscription.add(
      this.oidcSecurityService.getAccessToken().subscribe((accessToken) => {
        this.accessToken.set(accessToken);
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
    this.userInfoSubscription.unsubscribe();
    this.loginAttemptSubscription.unsubscribe();
    this.updateDialogSubscription.unsubscribe();
    this.deleteDialogSubscription.unsubscribe();
    this.newDialogSubscription.unsubscribe();
  }

  openDialog() {
    const dialogRef = this.dialog.open(SessionDialogComponent);
    dialogRef.afterClosed().subscribe((sessionInfoDialog) => {
      if (sessionInfoDialog) {
        this.newDialogSubscription.add(
          this.sessionInfoService
            .newSession$(
              this.accessToken(),
              this.userInfo().ID,
              sessionInfoDialog,
            )
            .pipe(
              map(({ data, err }) => {
                if (err) {
                  throw err;
                } else if (!data) {
                  throw Error('cannot read data');
                } else {
                  return data;
                }
              }),
            )
            .subscribe({
              next: (newSessionInfo) => {
                const sessionId = newSessionInfo?.ID;
                this._snackBar.open(
                  `Create Session ${sessionId} Success!`,
                  'Dismiss',
                  {
                    duration: 3000,
                  },
                );
                this.displaySessionInfo.update((displaySessionInfos) => [
                  ...displaySessionInfos,
                  {
                    session_info: signal(newSessionInfo),
                    should_hidden: false,
                  },
                ]);
              },
              error: (err) => {
                console.error(err);
              },
            }),
        );
      }
    });
  }

  openEditDialog(id: number) {
    if (!this.displaySessionInfo()) {
      return;
    }
    if (id < 0 || id >= this.displaySessionInfo().length) {
      return;
    }

    const sessionId = this.displaySessionInfo()[id].session_info().ID;

    const dialogRef = this.dialog.open(SessionDialogComponent, {
      data: {
        id: sessionId,
        resources: JSON.parse(
          this.displaySessionInfo()[id].session_info().Resources,
        ),
      },
    });

    dialogRef.afterClosed().subscribe((sessionInfoDialog) => {
      if (sessionInfoDialog) {
        this._snackBar.open('Loading...', 'Dismiss');
        this.updateDialogSubscription.add(
          this.sessionInfoService
            .updateSession$(
              this.accessToken(),
              this.userInfo().ID,
              sessionInfoDialog,
            )
            .pipe(
              map(({ data, err }) => {
                if (err) {
                  throw err;
                } else if (!data) {
                  throw Error('cannot read data');
                } else {
                  return data;
                }
              }),
            )
            .subscribe({
              next: (updatedSessionInfo) => {
                this._snackBar.open(
                  `Update Session ${sessionId} Success!`,
                  'Dismiss',
                  {
                    duration: 3000,
                  },
                );
                this.displaySessionInfo()[id].session_info.set(
                  updatedSessionInfo,
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
            }),
        );
      }
    });
  }

  openDeleteDialog(id: number) {
    if (!this.displaySessionInfo()) {
      return;
    }
    if (id < 0 || id >= this.displaySessionInfo().length) {
      return;
    }

    const sessionId = this.displaySessionInfo()[id].session_info().ID;
    const dialogRef = this.dialog.open(YesNoDialogComponent);

    dialogRef.afterClosed().subscribe((userCollect) => {
      if (userCollect === true) {
        this._snackBar.open('Loading...', 'Dismiss');
        this.deleteDialogSubscription.add(
          this.sessionInfoService
            .deleteSession$(this.accessToken(), this.userInfo().ID, sessionId)
            .pipe(
              map(({ data, err }) => {
                if (err) {
                  throw err;
                } else if (!data) {
                  throw Error('cannot read data');
                } else {
                  return data;
                }
              }),
            )
            .subscribe({
              next: (deletedSession) => {
                this._snackBar.open(
                  `Delete Session ${sessionId} Success!`,
                  'Dismiss',
                  {
                    duration: 3000,
                  },
                );
                const displaySessionInfos = this.displaySessionInfo();
                displaySessionInfos[id] = {
                  ...displaySessionInfos[id],
                  should_hidden: true,
                };
                this.displaySessionInfo.update((_) => [...displaySessionInfos]);
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
            }),
        );
      }
    });
  }

  switchSession(id: number, event: MatSlideToggleChange) {
    if (!this.displaySessionInfo()) {
      return;
    }
    if (id < 0 || id >= this.displaySessionInfo().length) {
      return;
    }
    const sessionId = this.displaySessionInfo()[id].session_info().ID;
    const newSessionInfo: SessionDialogInfo = {
      id: sessionId,
      resources: JSON.parse(
        this.displaySessionInfo()[id].session_info().Resources,
      ),
    };
    this._snackBar.open('Loading...', 'Dismiss');
    this.updateDialogSubscription.add(
      this.sessionInfoService
        .updateSession$(
          this.accessToken(),
          this.userInfo().ID,
          newSessionInfo,
          event.checked,
        )
        .pipe(
          map(({ data, err }) => {
            if (err) {
              throw err;
            } else if (!data) {
              throw Error('cannot read data');
            } else {
              return data;
            }
          }),
        )
        .subscribe({
          next: (updatedSessionInfo) => {
            this._snackBar.open(
              `Update Session ${sessionId} Success!`,
              'Dismiss',
              {
                duration: 3000,
              },
            );
            this.displaySessionInfo()[id].session_info.set(updatedSessionInfo);
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
        }),
    );
  }
}
