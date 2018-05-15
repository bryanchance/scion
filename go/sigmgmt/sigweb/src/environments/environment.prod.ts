import { LogLevel } from '../app/api/http-interceptors/log-level'
import { PROD_URL } from './prod-url'

export const environment = {
  production: true,
  logLevel: LogLevel.Error,
  domain: PROD_URL,
  url: 'https://' + PROD_URL
}
