import { Component, OnInit } from '@angular/core'

import { ApiService } from '../api/api.service'
import { PathPolicyFile } from '../sites/models/models'

@Component({
  selector: 'ana-policies',
  templateUrl: './policies.component.html',
  styleUrls: ['./policies.component.scss']
})
export class PoliciesComponent implements OnInit {
  globalPolicies: PathPolicyFile[]
  sitePolicies: PathPolicyFile[]

  constructor(private api: ApiService) { }

  ngOnInit() {
    this.api.getPathPolicies().subscribe(
      policies => {
        this.globalPolicies = policies.filter(p => p.Type === 'global')
        this.sitePolicies = policies.filter(p => p.Type === 'site')
      }
    )
  }
}
