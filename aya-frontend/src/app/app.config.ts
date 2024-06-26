import {
  HttpClientModule,
  provideHttpClient,
  withFetch,
} from '@angular/common/http';
import { ApplicationConfig } from '@angular/core';
import { provideClientHydration } from '@angular/platform-browser';
import { provideFileRouter } from '@analogjs/router';
import {
  provideRouter,
  Routes,
  withComponentInputBinding,
} from '@angular/router';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideAuth } from 'angular-auth-oidc-client';
import { authConfig } from './auth/auth.config';

const appRoutes: Routes = [{ path: '', pathMatch: 'full', redirectTo: 'home' }];
export const appConfig: ApplicationConfig = {
  providers: [
    provideFileRouter(),
    provideHttpClient(withFetch()),
    provideClientHydration(),
    provideFileRouter(withComponentInputBinding()),
    provideAnimations(),
    provideAuth(authConfig),
    provideRouter(appRoutes),
  ],
};
