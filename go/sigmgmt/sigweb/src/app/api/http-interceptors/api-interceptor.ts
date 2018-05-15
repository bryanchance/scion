import { HttpHandler, HttpInterceptor, HttpRequest } from '@angular/common/http'
import { Injectable } from '@angular/core'

import { environment } from '../../../environments/environment'

export const backendUrl = environment.url + '/api/'

@Injectable()
export class ApiInterceptor implements HttpInterceptor {
    intercept(req: HttpRequest<any>, next: HttpHandler) {
        if (req.url.search('doc') !== -1) {
            return next.handle(req)
        }

        const authReq = req.clone({
            headers: req.headers.set('Accept', 'application/json').set('Content-Type', 'application/json'),
            url: backendUrl + req.url
        })
        return next.handle(authReq)
    }
}
