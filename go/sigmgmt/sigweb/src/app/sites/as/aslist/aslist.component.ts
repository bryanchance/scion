import { Component, Input, OnInit, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../../api/api.service'
import { UserService } from '../../../api/user.service'
import { ASEntry, Site } from '../../models/models'
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'ana-aslist',
  templateUrl: './aslist.component.html',
  styleUrls: ['./aslist.component.scss']
})
export class ASListComponent implements OnInit {
  site: Site
  ia = new ASEntry
  ias: ASEntry[] = []
  success = ''
  error = ''
  editing = false
  @ViewChild('iaForm') form: NgForm

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    private userService: UserService) { }

  ngOnInit() {
    const siteID = this.route.snapshot.params.site

    if (siteID) {
      this.api.getSite(siteID).subscribe(
        site => {
          this.site = site
          this.api.getASes(this.site).subscribe(
            (ias: ASEntry[]) => this.ias = ias
          )
        }
      )
    }
  }

  onSubmit() {
    this.error = ''
    if (this.editing) {
      this.api.updateAS(this.site, this.ia).subscribe(
        ia => {
          this.ia = new ASEntry
          this.form.resetForm()
          this.editing = false
          this.success = 'Successfully updated AS.'
        },
        error => this.error = error.msg
      )
    } else {
      this.api.createAS(this.site, this.ia).subscribe(
        ia => {
          this.ias.push(ia)
          this.form.resetForm()
        },
        error => this.error = error.msg
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.ia = this.ias[idx]
  }

  delete(idx: number) {
    this.api.deleteAS(this.site, this.ias[idx]).subscribe(
      () => this.ias.splice(idx, 1)
    )
  }

  get formDisabled() {
    return this.userService.level === 0
  }
}
