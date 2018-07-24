import { Component, OnDestroy, OnInit } from '@angular/core'
import { MatSnackBar } from '@angular/material'
import { ActivatedRoute, Router } from '@angular/router'
import { Subscription } from 'rxjs'
import { debounceTime } from 'rxjs/operators'

import { ApiService } from '../../api/api.service'
import { CondAllOf, CondAnyOf, CondBool, CondIPv4, CondNot, MatchDestination, MatchDSCP, MatchSource } from '../models/cond'
import { Site, TrafficClass } from '../models/models'
import { TrafficClassService } from './traffic-class.service'

@Component({
  selector: 'ana-traffic-class-details',
  templateUrl: './traffic-class-details.component.html',
  styleUrls: ['./traffic-class-details.component.scss']
})
export class TrafficClassDetailsComponent implements OnInit, OnDestroy {
  trClass = new TrafficClass()
  saver: Subscription
  siteID: string
  error = ''
  rawEditing = false
  saving = false
  saved = false

  constructor(
    private api: ApiService,
    private tcService: TrafficClassService,
    private snackBar: MatSnackBar,
    private route: ActivatedRoute,
    private router: Router) {
    this.saver = this.tcService.saver.subscribe(
      () => {
        if (this.trClass.ID) {
          this.saving = true
          this.saved = false
          this.api.updateTrafficClass(this.trClass).subscribe(
            () => {
              this.showMsg()
            },
            error => this.showError(error)
          )
        } else {
          this.create()
        }
      }
    )
  }

  ngOnInit() {
    this.tcService.tcID = -1
    this.siteID = this.route.snapshot.params.site
    this.tcService.site.next(new Site(this.siteID))

    const classID = this.route.snapshot.params.class
    if (classID) {
      this.api.getTrafficClass(classID).subscribe(
        tc => {
          this.trClass = tc
          this.tcService.tcID = tc.ID
        }
      )
    }
  }

  ngOnDestroy() {
    this.saver.unsubscribe()
  }

  onSubmit(rawToggle = false) {
    if (this.trClass.ID) {
      this.saving = true
      this.saved = false
      this.api.updateTrafficClass(this.trClass).subscribe(
        tc => {
          this.trClass = tc
          this.showMsg()
          if (rawToggle) {
            this.rawEditing = false
          }
          if (this.rawEditing) {
            this.trClass.CondStr = this.trClass.condString
          }
        },
        error => this.showError(error)
      )
    } else {
      this.create()
    }
  }

  showMsg() {
    this.error = ''
    setTimeout(() => {
      this.saved = true
      this.saving = false
    }, 800)
  }

  showError(error) {
    this.error = error.msg
    this.saving = false
  }

  create() {
    this.api.createTrafficClass(new Site(this.siteID), this.trClass).subscribe(
      tc => this.router.navigate(['sites/' + this.siteID + '/classes/' + tc.ID]),
      error => this.showError(error)
    )
  }

  enableRawEdit() {
    this.trClass.CondStr = this.trClass.condString
    this.rawEditing = true
  }

  insert(condType: string) {
    switch (condType) {
      case 'all': {
        const all = new CondAllOf()
        all.Conds.push(this.trClass.Cond)
        this.trClass.Cond = all
        break
      }
      case 'any': {
        const any = new CondAnyOf()
        any.Conds.push(this.trClass.Cond)
        this.trClass.Cond = any
        break
      }
      case 'not': {
        const not = new CondNot()
        not.Operand = this.trClass.Cond
        this.trClass.Cond = not
      }
    }
  }

  addAny() {
    this.trClass.Cond = new CondAnyOf()
  }

  addAll() {
    this.trClass.Cond = new CondAllOf()
  }

  addNot() {
    this.trClass.Cond = new CondNot()
  }

  addBool() {
    this.trClass.Cond = new CondBool()
  }

  addSrc() {
    const predicate = new MatchSource()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.trClass.Cond = condIPv4
  }

  addDst() {
    const predicate = new MatchDestination()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.trClass.Cond = condIPv4
  }

  addDscp() {
    const predicate = new MatchDSCP()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.trClass.Cond = condIPv4
  }

  delete() {
    delete this.trClass.Cond
  }
}
