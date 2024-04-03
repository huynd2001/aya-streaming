import { PassedInitialConfig } from 'angular-auth-oidc-client';

import { environment } from '../../environments/environment';

export const authConfig: PassedInitialConfig = {
  config: {
    authority: environment.authority,
    redirectUrl: environment.redirectUrl,
    postLogoutRedirectUri: environment.redirectUrl,
    clientId: environment.clientId,
    scope: 'openid profile email offline_access', // 'openid profile offline_access ' + your scopes
    responseType: 'code',
    silentRenew: true,
    useRefreshToken: true,
    ignoreNonceAfterRefresh: true,
    triggerRefreshWhenIdTokenExpired: false,
    renewTimeBeforeTokenExpiresInSeconds: 30,
  },
};
