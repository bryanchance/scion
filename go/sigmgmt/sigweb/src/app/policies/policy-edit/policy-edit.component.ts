import { Component, OnInit } from '@angular/core'
import { ActivatedRoute, Router } from '@angular/router'
import * as yaml from 'js-yaml'
import { MarkdownService } from 'ngx-markdown'

import { ApiService } from '../../api/api.service'
import { PathPolicyFile } from '../../sites/models/models'


@Component({
  selector: 'ana-policy-edit',
  templateUrl: './policy-edit.component.html',
  styleUrls: ['./policy-edit.component.scss']
})
export class PolicyEditComponent implements OnInit {
  options = {}
  policies: any[]
  policy: PathPolicyFile = new PathPolicyFile()
  code = ''
  error = ''
  success = ''

  samplePolicy = '\
  - extends_example:\
      extends:\
      - sub_pol_1\
      - sub_pol_2\
  - sub_pol_1:\
      acl:\
      - "- 1-ff00:0:133#0"\
      - "+"\
  - sub_pol_2:\
      sequence: "0+ 1-ff00:0:110#0 1-ff00:0:110#0 0+"'

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    private router: Router,
    private markdownService: MarkdownService) {
  }

  ngOnInit() {
    this.markdownService.renderer.link = (href: string, title: string, text: string) => {
      return '<a href="config' + href + '">' + text + '</a>'
    }

    const policyID = this.route.snapshot.params.policy

    if (policyID) {
      this.api.getPathPolicy(policyID).subscribe(
        policy => this.setPolicy(policy),
        error => this.error = error.msg
      )
    }
  }

  setPolicy(policy: PathPolicyFile) {
    this.policy = policy
    this.code = yaml.safeDump(policy.Code)
  }

  save() {
    if (!this.check()) return
    if (this.policy.ID) {
      this.api.updatePathPolicy(this.policy).subscribe(
        () => this.setSuccess('Policy updated!'),
        error => this.setError(error.msg)
      )
    } else {
      this.api.createPathPolicy(this.policy).subscribe(
        policy => {
          this.setPolicy(policy)
          this.router.navigate(['/policies', this.policy.ID])
          this.setSuccess('Policy created!')
        },
        error => this.setError(error.msg)
      )
    }
  }

  check() {
    try {
      this.policy.Code = yaml.safeLoad(this.code)
    } catch (error) {
      this.setError(error.name + ': ' + error.message)
      return false
    }
    return this.api.validatePathPolicy(this.policy).subscribe(
      policy => {
        this.code = yaml.safeDump(policy.Code)
        this.setSuccess('Policy successfully validated.')
        return true
      },
      error => {
        if (error.desc.Field) {
          this.setError(error.msg + ' at ' + error.desc.Field)
        } else if (error.desc.Msg) {
          this.setError(error.desc.Msg)
        }
      }
    )
  }

  setError(error: string) {
    this.success = ''
    this.error = error
  }

  setSuccess(success: string) {
    this.error = ''
    this.success = success
  }
}
