import { Component, OnInit } from '@angular/core'
import { ActivatedRoute } from '@angular/router'
import { forkJoin } from 'rxjs'

import { ApiService } from '../../../api/api.service'
import { ASEntry, SIG, Site } from '../../models/models'

@Component({
  selector: 'ana-asdetail',
  templateUrl: './asdetail.component.html',
  styleUrls: ['./asdetail.component.scss']
})
export class ASDetailComponent implements OnInit {
  site: Site
  ia = new ASEntry
  sigs: SIG[]

  constructor(
    private route: ActivatedRoute,
    private api: ApiService
  ) { }

  ngOnInit() {
    const siteID = this.route.snapshot.params.site
    const asID = this.route.snapshot.params.ia

    if (siteID && asID) {
      forkJoin(
        this.api.getSite(siteID),
        this.api.getAS(asID)
      ).subscribe(
        ([site, ia]) => {
          this.site = site
          this.ia = ia
        },
        () => { }
      )
    }
  }
}
