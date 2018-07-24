import { Component, OnInit } from '@angular/core'
import { ActivatedRoute } from '@angular/router'

import { ApiService } from '../../api/api.service'
import { Site } from '../models/models'
import { TrafficClass } from '../models/models'

@Component({
  selector: 'ana-traffic-classes',
  templateUrl: './traffic-classes.component.html',
  styleUrls: ['./traffic-classes.component.scss']
})
export class TrafficClassesComponent implements OnInit {
  site: Site
  trafficClasses: TrafficClass[]

  constructor(private api: ApiService, private route: ActivatedRoute) { }

  ngOnInit() {
    const siteID = this.route.snapshot.params.site

    if (siteID) {
      this.api.getSite(siteID).subscribe(
        site => {
          this.site = site
          this.api.getTrafficClasses(this.site).subscribe(
            (classes: TrafficClass[]) => this.trafficClasses = classes
          )
        }
      )
    }
  }

  delete(e, idx: number) {
    e.preventDefault()
    e.stopPropagation()
    this.api.deleteTrafficClass(this.site, this.trafficClasses[idx])
      .subscribe(
        () => this.trafficClasses.splice(idx, 1),
        () => { }
      )
  }
}
