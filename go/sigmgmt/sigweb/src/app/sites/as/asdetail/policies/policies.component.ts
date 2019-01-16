import { Component, Input, OnChanges, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'
import { forkJoin } from 'rxjs'

import { ApiService } from '../../../../api/api.service'
import { UserService } from '../../../../api/user.service'
import { ASEntry, TrafficPolicy, Site, TrafficClass, PathPolicyFile, PathPolicy } from '../../../models/models'

@Component({
  selector: 'ana-policies',
  templateUrl: './policies.component.html',
  styleUrls: ['./policies.component.scss']
})
export class PoliciesComponent implements OnChanges {
  @Input() ia: ASEntry
  @Input() site: Site
  @ViewChild('trafficPolicyForm') form: NgForm
  success = ''
  error = ''
  editing = false

  trafficPolicies: TrafficPolicy[]
  trafficPolicy = new TrafficPolicy
  trafficClasses: TrafficClass[]
  searchPP = ''
  pathPolicies: PathPolicy[] = []
  filteredTCs: TrafficClass[]
  filteredPathPolicies: PathPolicy[]

  constructor(private api: ApiService, private userService: UserService) { }

  onTCChange(val) {
    this.filteredTCs = val === '' ? this.trafficClasses : this.trafficClasses.filter(option =>
      option.Name.toLowerCase().includes(val.toLowerCase()))
  }

  searchPathPolicy() {
    // Remove selectors that are already used, then filter for search text
    const pathPolicies = this.pathPolicies.filter(el => this.trafficPolicy.PathPolicies.indexOf(el.Name) === -1)
    this.filteredPathPolicies = this.searchPP === '' ? pathPolicies : pathPolicies.filter(option =>
      option.Name.toLowerCase().includes(this.searchPP.toLowerCase()))
  }

  addPathPolicy(name: string) {
    const pathPolicy = this.pathPolicies.find(el => el.Name === name)
    this.trafficPolicy.PathPolicies.push(pathPolicy.Name)
    this.searchPathPolicy()
  }

  removePathPolicy(idx: number) {
    this.trafficPolicy.PathPolicies.splice(idx, 1)
    this.searchPathPolicy()
  }

  getTrafficClass(id: number) {
    return this.trafficClasses.find(el => el.ID === id)
  }

  ngOnChanges() {
    if (this.ia && this.site) {
      forkJoin(
        this.api.getTrafficPolicies(this.ia),
        this.api.getTrafficClasses(this.site),
        this.api.getPathPolicies()
      ).subscribe(
        ([trPol, tcs, pPolFiles]) => {
          this.trafficPolicies = trPol
          this.trafficClasses = tcs
          // go through policy files
          pPolFiles.forEach((ppf) => {
            if (ppf.Code) {
              // go through policies in a policy file
              ppf.Code.forEach((pp2) => {
                const npp = new PathPolicy(Object.keys(pp2)[0], pp2[Object.keys(pp2)[0]])
                this.pathPolicies.push(npp)
              })
            }
          })
          this.filteredPathPolicies = this.pathPolicies
        },
        error => this.error = error.msg
      )
    }
  }

  onSubmit() {
    this.clearMsg()
    if (this.editing) {
      this.api.updatePolicy(this.ia, this.trafficPolicy).subscribe(
        policy => {
          this.trafficPolicy = new TrafficPolicy
          this.form.resetForm()
          this.editing = false
          this.success = 'Successfully updated Policy.'
        },
        error => this.error = error.msg
      )
    } else {
      this.api.createPolicy(this.ia, this.trafficPolicy).subscribe(
        policy => {
          this.trafficPolicies.push(policy)
          this.form.resetForm()
          this.trafficPolicy.PathPolicies = []
          this.searchPathPolicy()
        },
        error => this.error = error.msg
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.trafficPolicy = this.trafficPolicies[idx]
    this.searchPathPolicy()
  }

  delete(idx: number) {
    this.clearMsg()
    this.api.deletePolicy(this.trafficPolicies[idx]).subscribe(
      () => this.trafficPolicies.splice(idx, 1)
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
