import { Component, Input, OnChanges } from '@angular/core'

import { ApiService } from '../../../../api/api.service'
import { ASEntry } from '../../../models'

@Component({
  selector: 'ana-classifiers',
  templateUrl: './classifiers.component.html',
  styleUrls: ['./classifiers.component.scss']
})
export class ClassifiersComponent implements OnChanges {
  @Input() ia: ASEntry
  success = ''
  error = ''

  constructor(private api: ApiService) { }

  ngOnChanges() {
    if (this.ia.ID) {
      this.api.getAS(this.ia.ID).subscribe(
        (ia: ASEntry) => this.ia = ia,
        () => { }
      )
    }
  }

  onSubmit() {
    this.success = ''
    this.error = ''
    this.api.updateASPolicies(this.ia).subscribe(
      () => this.success = 'Policies successfully updated!',
      error => this.error = error.msg
    )
  }
}
