import { Component, OnInit } from '@angular/core'
import { ActivatedRoute } from '@angular/router'

import { ApiService } from '../../api/api.service'
import { Site } from '../models/models'

@Component({
  selector: 'ana-site-details',
  templateUrl: './site-details.component.html',
  styleUrls: ['./site-details.component.scss']
})
export class SiteDetailsComponent implements OnInit {
  site: Site
  newSite = true
  reloadSuccess = false
  loadingConfig = false
  reloadError = ''
  reloadErrorDesc: any
  errorKeys: string[]

  constructor(
    private route: ActivatedRoute,
    private api: ApiService) { }

  ngOnInit() {
    const id = this.route.snapshot.params.site
    if (id) {
      this.api.getSite(id).subscribe(
        site => {
          this.site = site
          this.newSite = false
        }
      )
    }
  }

  reloadConfig() {
    this.reloadError = ''
    this.reloadSuccess = false
    this.loadingConfig = true
    this.api.reloadConfig(this.site.ID).subscribe(
      () => {
        this.reloadSuccess = true
        this.loadingConfig = false
      },
      error => {
        this.reloadError = error.msg
        this.errorKeys = Object.keys(error.desc).filter(k =>
          k === this.site.VHost || this.site.Hosts.map(host => host.Name).includes(k))
        this.reloadErrorDesc = error.desc
        this.loadingConfig = false
      }
    )
  }
}
