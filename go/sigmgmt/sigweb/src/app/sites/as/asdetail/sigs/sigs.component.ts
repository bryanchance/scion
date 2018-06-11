import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../../../api/api.service'
import { UserService } from '../../../../api/user.service'
import { ASEntry, SIG, Site } from '../../../models'

@Component({
  selector: 'ana-sigs',
  templateUrl: './sigs.component.html',
  styleUrls: ['./sigs.component.scss']
})
export class SigsComponent implements OnChanges {
  @Input() ia: ASEntry
  success = ''
  error = ''
  editing = false

  sig = new SIG
  defaultEncapPort: number
  defaultCtrlPort: number
  sigs: SIG[]
  @ViewChild('sigForm') form: NgForm

  constructor(
    private api: ApiService,
    private userService: UserService
  ) {
    this.api.getDefaultSIG().subscribe(
      (sig: SIG) => {
        this.defaultCtrlPort = sig.CtrlPort
        this.defaultEncapPort = sig.EncapPort
      },
      () => { }
    )
  }

  ngOnChanges() {
    if (this.ia.ID) {
      this.api.getSIGs(this.ia).subscribe(
        sigs => {
          this.sigs = sigs
          this.setDefaultPorts()
        }
      )
    }
  }

  onSubmit() {
    this.clearMsg()
    if (this.editing) {
      this.api.updateSIG(this.sig).subscribe(
        sig => {
          this.sig = new SIG
          this.form.resetForm()
          this.setDefaultPorts()
          this.editing = false
          this.success = 'Successfully updated SIG.'
        },
        error => this.error = error.msg
      )
    } else {
      this.api.createSIG(this.ia, this.sig).subscribe(
        sig => {
          this.sigs.push(sig)
          this.form.resetForm()
          this.setDefaultPorts()
        },
        error => this.error = error.msg
      )
    }
  }

  setDefaultPorts() {
    this.form.controls['encapPort'].setValue(this.defaultEncapPort)
    this.form.controls['ctrlPort'].setValue(this.defaultCtrlPort)
  }

  edit(idx: number) {
    this.editing = true
    this.sig = this.sigs[idx]
  }

  delete(idx: number) {
    this.clearMsg()
    this.api.deleteSIG(this.sigs[idx]).subscribe(
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
