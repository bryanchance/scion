import { ChangeDetectorRef, Component, EventEmitter, Input, OnChanges, Output, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'
import { Router } from '@angular/router'

import { ApiService } from '../../api/api.service'
import { UserService } from '../../api/user.service'
import { Host, Site } from '../models'

@Component({
  selector: 'ana-site-configuration',
  templateUrl: './site-configuration.component.html',
  styleUrls: ['./site-configuration.component.scss']
})
export class SiteConfigurationComponent implements OnChanges {
  @Input() site: Site
  @Input() newSite: boolean
  @ViewChild('hostForm') hostForm: NgForm

  host = new Host
  success = ''
  error = ''
  editing = false

  constructor(
    private api: ApiService,
    private userService: UserService,
    private cd: ChangeDetectorRef,
    private router: Router
  ) { }

  ngOnChanges() {
    this.cd.detectChanges()
  }

  onSubmit() {
    this.clearMsg()
    if (this.newSite) {
      this.api.createSite(this.site).subscribe(
        site => {
          this.site = site
          this.newSite = false
          this.success = 'Successfully created Site.'
          this.router.navigate(['/sites', this.site.ID])
        },
        error => this.error = error.msg
      )
    } else {
      this.api.updateSite(this.site).subscribe(
        site => {
          this.site = site
          this.success = 'Successfully updated Site.'
        },
        error => this.error = error.msg
      )
    }
  }

  saveHost() {
    this.clearMsg()
    const site = { ...this.site }
    if (!this.editing) {
      const hosts = Object.assign([], this.site.Hosts)
      hosts.push({ ...this.host })
      site.Hosts = hosts
    }
    this.api.updateSite(site).subscribe(
      () => {
        this.success = 'Successfully updated Hosts.'
        this.host = new Host
        this.hostForm.resetForm()
        this.editing = false
        this.site = site
      },
      error => {
        this.error = error.msg
      }
    )
  }

  editHost(idx: number) {
    this.editing = true
    this.host = this.site.Hosts[idx]
  }

  deleteHost(idx: number) {
    this.clearMsg()
    this.site.Hosts.splice(idx, 1)
    this.api.updateSite(this.site).subscribe(
      () => this.success = 'Successfully updated Hosts.',
      error => this.error = error.msg
    )
  }

  get formDisabled() {
    return this.userService.level === 0
  }

  clearMsg() {
    this.success = ''
    this.error = ''
  }
}
