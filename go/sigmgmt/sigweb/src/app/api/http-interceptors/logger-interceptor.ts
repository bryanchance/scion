import { HttpErrorResponse, HttpHandler, HttpInterceptor, HttpRequest, HttpResponse } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { throwError } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { environment } from '../../../environments/environment'
import { UserService } from '../user.service'
import { LogLevel } from './log-level'

export class AnaError {
    constructor(public msg: string, public desc: any = '') { }
}

@Injectable()
export class LoggerInterceptor implements HttpInterceptor {

    constructor(private userService: UserService) { }

    logLevel = environment.logLevel

    intercept(req: HttpRequest<any>, next: HttpHandler) {
        return next.handle(req).pipe(
            map(resp => {
                if (resp instanceof HttpResponse) {
                    this.debug('Response', resp)
                    return resp
                }
            }),
            catchError(err => {
                this.error('error', err.error)
                if (err instanceof HttpErrorResponse) {
                    this.error('HttpErrorResponse', err.status)
                    if (err.status === 401) {
                        this.userService.logout()
                    }
                }
                if (err.error && err.error.error) {
                    if (err.error.description) {
                        return throwError(new AnaError(err.error.error, err.error.description))
                    }
                    return throwError(new AnaError(err.error.error))
                } else {
                    return throwError(new AnaError('Something went wrong!'))
                }
            })
        )
    }

    debug(msg: string, ...obj: any[]): void {
        if (this.logLevel <= LogLevel.Debug) {
            // tslint:disable-next-line
            console.debug('[Debug] ' + msg, ...obj)
        }
    }

    error(msg: string, ...obj: any[]): void {
        if (this.logLevel <= LogLevel.Error) {
            console.error('[Error] ' + msg, ...obj)
        }
    }

    info(msg: string, ...obj: any[]): void {
        if (this.logLevel <= LogLevel.Info) {
            // tslint:disable-next-line
            console.info('[Info] ' + msg, ...obj)
        }
    }

    warn(msg: string, ...obj: any[]): void {
        if (this.logLevel <= LogLevel.Warn) {
            console.warn('[Warning] ' + msg, ...obj)
        }
    }

}
