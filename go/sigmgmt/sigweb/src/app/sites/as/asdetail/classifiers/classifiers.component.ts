import { Component, Input, OnChanges } from '@angular/core'

import { ApiService } from '../../../../api/api.service'
import { IA, Policy, Site } from '../../../models'

@Component({
  selector: 'ana-classifiers',
  templateUrl: './classifiers.component.html',
  styleUrls: ['./classifiers.component.scss']
})
export class ClassifiersComponent implements OnChanges {
  @Input() site: Site
  @Input() ia: IA
  success = ''
  error = ''

  constructor(private api: ApiService) { }

  ngOnChanges() {
    if (this.site.Name && this.ia) {
      this.api.getIA(this.site, this.ia).subscribe(
        (ia: IA) => this.ia = ia
      )
    }
  }

  onSubmit() {
    this.success = ''
    this.error = ''
    this.api.updateIAPolicies(this.site, this.ia, this.ia.Policy).subscribe(
      () => this.success = 'Policies successfully updated!',
      error => this.error = error
    )
  }
}
