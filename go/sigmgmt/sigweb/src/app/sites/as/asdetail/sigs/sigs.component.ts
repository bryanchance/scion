import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../../../api/api.service'
import { UserService } from '../../../../api/user.service'
import { IA, SIG, Site } from '../../../models'

@Component({
  selector: 'ana-sigs',
  templateUrl: './sigs.component.html',
  styleUrls: ['./sigs.component.scss']
})
export class SigsComponent implements OnChanges {
  @Input() site: Site
  @Input() ia: IA
  success = ''
  error = ''
  editing = false

  sig = new SIG
  sigs: SIG[]
  @ViewChild('sigForm') form: NgForm

  constructor(private api: ApiService, private userService: UserService) { }

  ngOnChanges() {
    if (this.site.Name && this.ia) {
      this.api.getSIGs(this.site, this.ia).subscribe(
        sigs => this.sigs = sigs
      )
    }
  }

  onSubmit() {
    this.clearMsg()
    if (this.editing) {
      this.api.updateSIG(this.site, this.ia, this.sig).subscribe(
        sig => {
          this.sig = new SIG
          this.form.resetForm()
          this.editing = false
          this.success = 'Successfully updated SIG.'
         },
        error => this.error = error
      )
    } else {
      this.api.createSIG(this.site, this.ia, this.sig).subscribe(
        sig => {
          this.sigs.push({ ...sig })
          this.form.resetForm()
        },
        error => this.error = error
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.sig = this.sigs[idx]
  }

  delete(idx: number) {
    this.clearMsg()
    this.api.deleteSIG(this.site, this.ia, this.sigs[idx]).subscribe(
      () => this.sigs.splice(idx, 1)
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
