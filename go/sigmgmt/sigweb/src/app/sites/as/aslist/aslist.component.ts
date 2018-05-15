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
    this.api.createIA(this.site, this.ia).subscribe(
      ia => {
        this.ias.push({ ...ia })
        this.form.resetForm()
      },
      error => this.error = error
    )
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
