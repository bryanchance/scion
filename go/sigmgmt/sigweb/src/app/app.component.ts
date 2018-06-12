import { Component } from '@angular/core'
import { Router } from '@angular/router'

import { UserService } from './api/user.service'
import { MatDialog, MatDialogRef } from '@angular/material'
import { ApiService } from './api/api.service'

@Component({
  selector: 'ana-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  dialogRef: MatDialogRef<OfflineDialogComponent, {}>
  timer: any

  constructor(public userService: UserService, public dialog: MatDialog, private api: ApiService) {
    this.userService.online.subscribe(
      online => {
        if (this.dialogRef && online) {
          this.dialogRef.close()
          clearTimeout(this.timer)
          delete (this.dialogRef)
          window.location.reload()
        }
        if (!online && !this.dialogRef) {
          this.dialogRef = this.dialog.open(OfflineDialogComponent, { disableClose: true })
          this.reconnect()
        }
      }
    )
  }

  logout() {
    this.userService.logout()
  }

  reconnect() {
    this.timer = setInterval(() => {
      this.api.getSites().subscribe()
    }, 2000)
  }
}

@Component({
  selector: 'ana-offline-dialog',
  template: `<span class="mat-typography">We can't reach the API at the moment. Trying to reconnect...</span>`,
})
export class OfflineDialogComponent {
  constructor() { }
}
