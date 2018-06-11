import { Component, Input, OnInit, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'

import { ApiService } from '../../api/api.service'
import { UserService } from '../../api/user.service'
import { PathSelector, Site } from '../models'

@Component({
  selector: 'ana-path-selector',
  templateUrl: './path-selector.component.html',
  styleUrls: ['./path-selector.component.scss']
})
export class PathSelectorComponent implements OnInit {
  @Input() site: Site
  pathSelector = new PathSelector
  pathSelectors: PathSelector[]
  success = ''
  error = ''
  editing = false
  @ViewChild('pathForm') form: NgForm

  constructor(private api: ApiService, private userService: UserService) { }

  ngOnInit() {
    this.api.getPathSelectors(this.site).subscribe(
      ps => this.pathSelectors = ps,
      () => { }
    )
  }

  onSubmit() {
    this.error = ''
    if (this.editing) {
      this.api.updatePathSelector(this.pathSelector).subscribe(
        () => {
          this.pathSelector = new PathSelector
          this.form.resetForm()
          this.editing = false
          this.success = 'Successfully updated PathSelector.'
        },
        error => this.error = error.msg
      )
    } else {
      this.api.createPathSelector(this.site, this.pathSelector).subscribe(
        selector => {
          this.pathSelectors.push(selector)
          this.form.resetForm()
        },
        error => this.error = error.msg
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.pathSelector = this.pathSelectors[idx]
  }

  delete(idx: number) {
    this.error = ''
    this.api.deletePathSelector(this.site, this.pathSelectors[idx])
      .subscribe(
        () => this.pathSelectors.splice(idx, 1),
        (error) => this.error = error
      )
  }

  get formDisabled() {
    return this.userService.level === 0
  }
}
