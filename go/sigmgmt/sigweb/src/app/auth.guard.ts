import { Injectable } from '@angular/core'
import { ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot } from '@angular/router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { UserService } from './api/user.service'

@Injectable()
export class AuthGuard implements CanActivate {
  constructor(
    private router: Router,
    private userService: UserService
  ) { }

  canActivate(
    next: ActivatedRouteSnapshot,
    state: RouterStateSnapshot): Observable<boolean> | Promise<boolean> | boolean {


    return this.userService.isLoggedIn.pipe(
      map(loggedIn => {
        if (loggedIn) {
          return true
        } else {
          this.router.navigate(['/login'])
          return false
        }
      })
    )
  }
}
