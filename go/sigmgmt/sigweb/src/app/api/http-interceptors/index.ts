import { HTTP_INTERCEPTORS } from '@angular/common/http'

import { ApiInterceptor } from './api-interceptor'
import { LoggerInterceptor } from './logger-interceptor'

export const httpInterceptorProviders = [
  { provide: HTTP_INTERCEPTORS, useClass: ApiInterceptor, multi: true },
  { provide: HTTP_INTERCEPTORS, useClass: LoggerInterceptor, multi: true },
]
