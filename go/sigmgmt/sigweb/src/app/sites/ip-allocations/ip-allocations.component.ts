import { Component, OnInit, ViewChild } from '@angular/core'
import { ApiService } from '../../api/api.service'
import { Site, SiteNetwork } from '../models/models'
import { ActivatedRoute } from '@angular/router'
import { NgForm } from '@angular/forms'

@Component({
  selector: 'ana-ip-allocations',
  templateUrl: './ip-allocations.component.html',
  styleUrls: ['./ip-allocations.component.scss']
})
export class IpAllocationsComponent implements OnInit {
  success = ''
  error = ''
  @ViewChild('allocationForm') form: NgForm
  editing = false
  site: Site
  allocation = new SiteNetwork
  allocations: SiteNetwork[]

  constructor(private api: ApiService, private route: ActivatedRoute) { }

  ngOnInit() {
    const siteID = this.route.snapshot.params.site

    if (siteID) {
      this.api.getSite(siteID).subscribe(
        site => {
          this.site = site
          this.api.getIPAllocations(this.site).subscribe(
            allocations => this.allocations = allocations
          )
        }
      )
    }
  }

  onSubmit() {
    this.clearMsg()
    if (this.editing) {
      this.api.updateIPAllocation(this.allocation).subscribe(
        () => {
          this.allocations.push(this.allocation)
          this.allocation = new SiteNetwork
          this.form.resetForm()
          this.editing = false
          this.setSuccess('Successfully updated IP Allocation.')
        },
        error => this.setError(error.msg)
      )
    } else {
      this.api.createIPAllocation(this.site, this.allocation).subscribe(
        allocation => {
          this.allocations.push(allocation)
          this.form.resetForm()
          this.setSuccess('Successfully created IP Allocation.')
        },
        error => this.setError(error.msg)
      )
    }
  }

  edit(idx: number) {
    this.editing = true
    this.allocation = this.allocations[idx]
    this.allocations.splice(idx, 1)
  }

  delete(idx: number) {
    this.clearMsg()
    this.api.deleteIPAllocation(this.allocations[idx]).subscribe(
      () => this.allocations.splice(idx, 1)
    )
  }

  clearMsg() {
    this.success = ''
    this.error = ''
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
