import { Component, OnInit } from '@angular/core'
import { ActivatedRoute } from '@angular/router'

import { ApiService } from '../../api/api.service'
import { Site } from '../models'

@Component({
  selector: 'ana-site-details',
  templateUrl: './site-details.component.html',
  styleUrls: ['./site-details.component.scss']
})
export class SiteDetailsComponent implements OnInit {
  site = new Site
  newSite = true
  reloadSuccess = false
  loadingConfig = false
  reloadError = ''

  constructor(
    private route: ActivatedRoute,
    private api: ApiService) { }

  ngOnInit() {
    const name = this.route.snapshot.paramMap.get('site')
    if (name) {
      this.api.getSite(name).subscribe(
        site => {
          this.site = site
          this.newSite = false
        }
      )
    }
  }

  reloadConfig() {
    this.reloadError = ''
    this.loadingConfig = true
    this.api.reloadConfig(this.site.Name).subscribe(
      () => {
        this.reloadSuccess = true
        this.loadingConfig = false
      },
      error => {
        this.reloadError = error
        this.loadingConfig = false
      }
    )
  }
}
