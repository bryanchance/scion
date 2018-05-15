import { Component } from '@angular/core'
import { Router } from '@angular/router'

import { User, UserService } from '../api/user.service'

@Component({
  selector: 'ana-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {

  user = new User
  error = ''

  constructor(private userService: UserService, private router: Router) { }

  onSubmit() {
    this.userService.login(this.user).subscribe(
      () => this.router.navigate(['/']),
      () => this.error = 'Bad Credentials!'
    )
  }

}
