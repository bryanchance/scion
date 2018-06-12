import { Component, EventEmitter, Input, Output } from '@angular/core'
import { Observable } from 'rxjs'
import { filter, map } from 'rxjs/operators'

import { ApiService } from '../../../../api/api.service'
import { CondClass } from '../../../models/cond'
import { TrafficClass } from '../../../models/models'
import { TrafficClassService } from '../../traffic-class.service'

@Component({
  selector: 'ana-class',
  templateUrl: './class.component.html',
  styleUrls: ['./class.component.scss']
})
export class ClassComponent {
  @Input() cond: CondClass
  trafficClasses$: Observable<TrafficClass[]>

  constructor(private tcService: TrafficClassService, private api: ApiService) {
    this.tcService.site.subscribe(
      site => this.trafficClasses$ = this.api.getTrafficClasses(site).pipe(
        map(tcs => tcs.filter(el => el.ID !== this.tcService.tcID))
      )
    )
  }

  onSubmit(e) {
    this.tcService.save.next(e.ID)
  }
}
