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
import { filter, Subscription } from 'rxjs';

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
  userDataSubscription: Subscription | undefined;
  authSubscription: Subscription | undefined;

  private readonly oidcSecurityService = inject(OidcSecurityService);
  private readonly eventService = inject(PublicEventsService);

  ngOnInit(): void {
    console.log('huh');
    this.eventService
      .registerForEvents()
      .pipe(
        filter(
          (event) =>
            event.type === EventTypes.SilentRenewStarted ||
            event.type === EventTypes.SilentRenewFailed
        )
      )
      .subscribe({
        next: (u) => {
          console.log(u);
        },
        error: (error) => {
          console.error(error);
        },
      });
    this.userDataSubscription = this.oidcSecurityService.userData$.subscribe({
      next: (userData) => {
        console.log(userData);
      },
      error: (error) => console.error(error),
    });
    this.authSubscription = this.oidcSecurityService.isAuthenticated$.subscribe(
      {
        next: (authRes) => {
          this.isAuth = authRes.isAuthenticated;
        },
        error: (error) => console.error(error),
      }
    );
    this.oidcSecurityService.checkAuth().subscribe((next) => {
      // Do nothing
    });
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
    this.userDataSubscription?.unsubscribe();
    this.authSubscription?.unsubscribe();
  }
}
