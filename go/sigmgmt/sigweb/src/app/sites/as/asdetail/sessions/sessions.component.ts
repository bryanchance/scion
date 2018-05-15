import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../../../api/api.service'
import { UserService } from '../../../../api/user.service'
import { IA, PathSelector, Session, Site } from '../../../models'

@Component({
  selector: 'ana-sessions',
  templateUrl: './sessions.component.html',
  styleUrls: ['./sessions.component.scss']
})
export class SessionsComponent implements OnChanges {
  @Input() site: Site
  @Input() ia: IA
  success = ''
  error = ''

  session = new Session
  sessions: Session[]
  sessionAliases: string[]
  pathSelectors: PathSelector[] = []
  @ViewChild('sessionForm') form: NgForm

  constructor(private api: ApiService, private userService: UserService) { }

  ngOnChanges(): void {
    if (this.site.Name && this.ia) {
      if (this.userService.level !== 0) {
        this.api.getPathSelectors(this.site).subscribe(
          pathSelectors => this.pathSelectors = pathSelectors
        )
        this.api.getSessions(this.site, this.ia).subscribe(
          sessions => this.sessions = sessions
        )
      } else {
        // get session aliases only
        this.api.getSessionAliases(this.site, this.ia).subscribe(
          sessionAliases => this.sessionAliases = sessionAliases
        )
      }
    }
  }

  onSubmit() {
    this.error = ''
    this.api.createSession(this.site, this.ia, this.session).subscribe(
      session => {
        this.sessions.push({ ...session })
        this.form.resetForm()
      },
      error => this.error = error
    )
  }

  delete(idx: number) {
    this.api.deleteSession(this.site, this.ia, this.sessions[idx]).subscribe(
      () => this.sessions.splice(idx, 1)
    )
  }

  getPP(name): string {
    return this.pathSelectors.filter(el => el.Name === name)[0].PP
  }

  get formDisabled() {
    return this.userService.level === 0
  }
}
