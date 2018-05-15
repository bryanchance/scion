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
  newPathSelector = new PathSelector
  pathSelectors: PathSelector[]
  success = ''
  error = ''
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
    this.api.createPathSelector(this.site, this.newPathSelector).subscribe(
      () => {
        this.pathSelectors.push({ ...this.newPathSelector })
        this.form.resetForm()
      },
      (error) => this.error = error
    )
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
