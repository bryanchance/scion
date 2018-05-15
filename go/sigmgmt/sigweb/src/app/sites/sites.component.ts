import { Component, OnDestroy, OnInit } from '@angular/core'

import { ApiService } from '../api/api.service'
import { UserService } from '../api/user.service'
import { Site } from './models'

@Component({
  selector: 'ana-sites',
  templateUrl: './sites.component.html',
  styleUrls: ['./sites.component.scss']
})
export class SitesComponent implements OnInit, OnDestroy {
  sites: Site[]

  constructor(
    private api: ApiService,
    private userService: UserService
  ) { }

  ngOnInit() {
    this.api.getSites().subscribe(
      sites => this.sites = sites
    )
  }

  ngOnDestroy() {
    this.api.getSites()
  }

  delete(idx: number) {
    this.api.deleteSite(this.sites[idx]).subscribe(
      () => this.sites.splice(idx, 1),
      () => { }
    )
  }

  get formDisabled() {
    return this.userService.level === 0
  }
}
