import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { MatToolbar } from '@angular/material/toolbar';
import { MatIcon } from '@angular/material/icon';
import { MatButton, MatIconButton } from '@angular/material/button';
import {
  EventTypes,
  OidcSecurityService,
  PublicEventsService,
} from 'angular-auth-oidc-client';
import { MatMenu, MatMenuItem, MatMenuTrigger } from '@angular/material/menu';
import { MatList, MatListItem } from '@angular/material/list';
import { filter, map, Subscription, switchMap, throwError } from 'rxjs';
import { UserInfoService } from '../services/user-info.service';
import { User } from '../interfaces/user';

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
  ],
  templateUrl: 'home.page.html',
  styleUrl: 'home.page.css',
})
export default class HomePage implements OnInit, OnDestroy {
  public isAuth: boolean = false;
  public userInfo: User | undefined;
  public isLoading: boolean = true;

  private authSubscription: Subscription = new Subscription();
  private userInfoSubscription: Subscription = new Subscription();
  private isLoadingSubscription: Subscription = new Subscription();

  private readonly oidcSecurityService = inject(OidcSecurityService);
  private readonly eventService = inject(PublicEventsService);
  private readonly userInfoService = inject(UserInfoService);

  ngOnInit(): void {
    let userInfo$ = this.oidcSecurityService
      .checkAuth()
      .pipe(
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
        })
      )
      .pipe(
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

    let isAuth$ = this.oidcSecurityService.isAuthenticated$.pipe(
      map((authResult) => authResult.isAuthenticated)
    );

    let isLoading$ = this.eventService.registerForEvents().pipe(
      filter(
        (event) =>
          event.type === EventTypes.CheckingAuth ||
          event.type === EventTypes.CheckingAuthFinished ||
          event.type === EventTypes.CheckingAuthFinishedWithError
      ),
      map((event) => {
        switch (event.type) {
          case EventTypes.CheckingAuth:
            return true;
          case EventTypes.CheckingAuthFinishedWithError:
            return false;
          case EventTypes.CheckingAuthFinished:
            return false;
          default:
            return true;
        }
      })
    );

    this.userInfoSubscription.add(
      userInfo$.subscribe({
        next: (userInfo) => {
          this.userInfo = userInfo;
        },
        error: (err) => {
          console.log('lmao');
          console.error(err);
        },
      })
    );

    this.authSubscription.add(
      isAuth$.subscribe({
        next: (isAuth) => {
          this.isAuth = isAuth;
        },
        error: (err) => {
          console.error(err);
        },
      })
    );

    this.isLoadingSubscription.add(
      isLoading$.subscribe({
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
  }
}
