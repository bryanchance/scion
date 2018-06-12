import { Injectable } from '@angular/core'
import { Router } from '@angular/router'
import { JwtHelperService } from '@auth0/angular-jwt'
import { BehaviorSubject, Observable } from 'rxjs'
import { tap } from 'rxjs/operators'

import { ApiService } from './api.service'

export class User {
  username: string
  password: string
}

class JWT {
  exp: number
  iat: number
  level: number
  sub: string
}

const TOKEN_NAME = 'jwt'

export function getToken(): string | null {
  return sessionStorage.getItem(TOKEN_NAME)
}

function storeToken(tokenString: string): void {
  sessionStorage.setItem(TOKEN_NAME, tokenString)
}

function removeToken(): void {
  sessionStorage.removeItem(TOKEN_NAME)
}

@Injectable()
export class UserService {

  online =  new BehaviorSubject<boolean>(true)
  private user: User
  private isLoginSubject = new BehaviorSubject<boolean>(this.token != null && !this.jwt.isTokenExpired())

  constructor(
    private api: ApiService,
    private jwt: JwtHelperService,
    private router: Router
  ) { }

  login(user: User): Observable<User> {
    this.logout()
    return this.api.obtainToken(user).pipe(
      tap((data: any) => {
        storeToken(data.token)
        this.user = user
        this.isLoginSubject.next(true)
      })
    )
  }

  logout(): void {
    removeToken()
    this.isLoginSubject.next(false)
    this.router.navigate(['/login'])
  }

  get isLoggedIn(): Observable<boolean> {
    return this.isLoginSubject.asObservable()
  }

  get token(): JWT | null {
    return this.jwt.decodeToken(getToken())
  }

  get level(): number {
    if (getToken()) {
      return this.token.level
    }
    return
  }
}
