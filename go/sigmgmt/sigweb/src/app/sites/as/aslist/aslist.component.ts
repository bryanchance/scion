import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../../api/api.service'
import { UserService } from '../../../api/user.service'
import { IA, Site } from '../../models'

@Component({
  selector: 'ana-aslist',
  templateUrl: './aslist.component.html',
  styleUrls: ['./aslist.component.scss']
})
export class ASListComponent implements OnChanges {
  @Input() site: Site
  ia = new IA
  ias: IA[] = []
  success = ''
  error = ''
  editing = false
  @ViewChild('iaForm') form: NgForm

  constructor(private api: ApiService, private userService: UserService) { }

  ngOnChanges() {
    if (this.site) {
      this.api.getIAs(this.site).subscribe(
        (ias: IA[]) => this.ias = ias
      )
    }
  }

  onSubmit() {
    this.error = ''
    if (this.editing) {
      this.api.updateIA(this.site, this.ia).subscribe(
        ia => {
          this.ia = new IA
          this.form.resetForm()
          this.editing = false
          this.success = 'Successfully updated AS.'
        },
        error => this.error = error
      )
    } else {
      this.api.createIA(this.site, this.ia).subscribe(
        ia => {
          this.ias.push({ ...ia })
          this.form.resetForm()
        },
        error => this.error = error
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.ia = this.ias[idx]
  }

  delete(idx: number) {
    this.api.deleteIA(this.site, this.ias[idx]).subscribe(
      () => this.ias.splice(idx, 1)
    )
  }

  get formDisabled() {
    return this.userService.level === 0
  }
}
