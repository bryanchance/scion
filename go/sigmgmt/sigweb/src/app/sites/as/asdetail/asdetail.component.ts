import { Component, OnInit } from '@angular/core'
import { ActivatedRoute } from '@angular/router'
import { forkJoin } from 'rxjs'

import { ApiService } from '../../../api/api.service'
import { IA, Session, SIG, Site } from '../../models'

@Component({
  selector: 'ana-asdetail',
  templateUrl: './asdetail.component.html',
  styleUrls: ['./asdetail.component.scss']
})
export class ASDetailComponent implements OnInit {
  site = new Site
  ia = new IA
  sessions: Session[]
  sigs: SIG[]

  constructor(
    private route: ActivatedRoute,
    private api: ApiService
  ) { }

  ngOnInit() {
    const siteParam = this.route.snapshot.params.site
    const iaParam = this.route.snapshot.params.ia

    if (siteParam && iaParam) {
      [this.ia.ISD, this.ia.AS] = iaParam.split('-', 2)
      forkJoin(
        this.api.getSite(siteParam),
        this.api.getIA(new Site(siteParam), this.ia)
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
