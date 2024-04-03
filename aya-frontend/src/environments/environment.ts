import { constants } from '../app/interfaces/constants';

export const environment = {
  production: true,
  wsSocketUrl: import.meta.env[constants.WEBSOCKET_URL_ENV] as string,
  redirectUrl: import.meta.env[constants.REDIRECT_URL_ENV] as string,
  clientId: import.meta.env[constants.AUTH_CLIENT_ID_ENV] as string,
  authority: import.meta.env[constants.AUTH_URL_ENV] as string,
};
