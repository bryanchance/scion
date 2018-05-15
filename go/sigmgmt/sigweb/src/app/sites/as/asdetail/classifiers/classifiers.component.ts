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

  policy = new Policy('')

  constructor(private api: ApiService) { }

  ngOnChanges() {
    if (this.site.Name && this.ia) {
      this.api.getIAPolicy(this.site, this.ia).subscribe(
        policy => this.policy = policy
      )
    }
  }

  onSubmit() {
    this.success = ''
    this.error = ''
    this.api.updateIAPolicy(this.site, this.ia, this.policy).subscribe(
      () => this.success = 'Policies successfully updated!',
      error => this.error = error
    )
  }
}
