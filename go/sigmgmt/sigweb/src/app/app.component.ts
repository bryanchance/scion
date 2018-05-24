import { Component } from '@angular/core'
import { Router } from '@angular/router'

import { UserService } from './api/user.service'

@Component({
  selector: 'ana-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {

  constructor(public userService: UserService) { }

  logout() {
    this.userService.logout()
  }
}
