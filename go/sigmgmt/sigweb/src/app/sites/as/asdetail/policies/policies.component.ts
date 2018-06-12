import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'
import { forkJoin } from 'rxjs'

import { ApiService } from '../../../../api/api.service'
import { UserService } from '../../../../api/user.service'
import { ASEntry, PathSelector, Policy, Site, TrafficClass } from '../../../models/models'


@Component({
  selector: 'ana-policies',
  templateUrl: './policies.component.html',
  styleUrls: ['./policies.component.scss']
})
export class PoliciesComponent implements OnChanges {
  @Input() ia: ASEntry
  @Input() site: Site
  @ViewChild('policyForm') form: NgForm
  success = ''
  error = ''
  editing = false

  policies: Policy[]
  policy = new Policy
  trafficClasses: TrafficClass[]
  searchSel = ''
  selectors: PathSelector[]
  filteredTCs: TrafficClass[]
  filteredSelectors: PathSelector[]

  constructor(private api: ApiService, private userService: UserService) { }

  onTCChange(val) {
    this.filteredTCs = val === '' ? this.trafficClasses : this.trafficClasses.filter(option =>
      option.Name.toLowerCase().includes(val.toLowerCase()))
  }

  searchSelector() {
    // Remove selectors that are already used, then filter for search text
    const selectors = this.selectors.filter(el => this.policy.Selectors.indexOf(el.ID) === -1)
    this.filteredSelectors = this.searchSel === '' ? selectors : selectors.filter(option =>
      option.Name.toLowerCase().includes(this.searchSel.toLowerCase()))
  }

  addSelector(id: number) {
    const selector = this.selectors.find(el => el.ID === id)
    this.policy.Selectors.push(selector.ID)
    this.searchSelector()
  }

  removeSelector(idx: number) {
    const selector = this.selectors.find(el => el.ID === this.policy.Selectors[idx])
    this.policy.Selectors.splice(idx, 1)
    this.searchSelector()
  }

  getSelector(id: number) {
    return this.selectors.find(el => el.ID === id)
  }

  getTrafficClass(id: number) {
    return this.trafficClasses.find(el => el.ID === id)
  }

  ngOnChanges() {
    if (this.ia && this.site) {
      forkJoin(
        this.api.getPolicies(this.ia),
        this.api.getTrafficClasses(this.site),
        this.api.getPathSelectors(this.site)
      ).subscribe(
        ([policies, tcs, sel]) => {
          this.policies = policies
          this.trafficClasses = tcs
          this.selectors = sel
          this.filteredSelectors = sel
        },
        error => this.error = error.msg
      )
    }
  }

  onSubmit() {
    this.clearMsg()
    if (this.editing) {
      this.api.updatePolicy(this.policy).subscribe(
        policy => {
          this.policy = new Policy
          this.form.resetForm()
          this.editing = false
          this.success = 'Successfully updated Policy.'
        },
        error => this.error = error.msg
      )
    } else {
      this.api.createPolicy(this.ia, this.policy).subscribe(
        policy => {
          this.policies.push(policy)
          this.form.resetForm()
          this.policy.Selectors = []
          this.searchSelector()
        },
        error => this.error = error.msg
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.policy = this.policies[idx]
  }

  delete(idx: number) {
    this.clearMsg()
    this.api.deletePolicy(this.policies[idx]).subscribe(
      () => this.policies.splice(idx, 1)
    )
  }

  clearMsg() {
    this.success = ''
    this.error = ''
  }


  get formDisabled() {
    return this.userService.level === 0
  }

}
