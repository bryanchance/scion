import { AfterViewInit, Component, Input, OnDestroy, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'
import { Subscription } from 'rxjs'
import { auditTime, debounceTime, distinctUntilChanged, filter, map, skip } from 'rxjs/operators'

import { MatchDestination } from '../../../models/cond'
import { TrafficClassService } from '../../traffic-class.service'

@Component({
  selector: 'ana-dst',
  templateUrl: './dst.component.html',
  styleUrls: ['./dst.component.scss']
})
export class DstComponent implements OnDestroy, AfterViewInit {
  @Input() predicate: MatchDestination
  @ViewChild('dstForm') form: NgForm
  sub: Subscription
  error = false

  constructor(public saveService: TrafficClassService) { }

  ngAfterViewInit() {
    this.sub = this.form.form.valueChanges.pipe(
      skip(1),
      map(el => el.name),
      debounceTime(500),
      filter(el => this.valid(el)),
      auditTime(400),
      distinctUntilChanged(),
      debounceTime(1000),
    ).subscribe(
      () => this.save()
    )
  }

  ngOnDestroy() {
    this.sub.unsubscribe()
  }

  save() {
    if (this.valid(this.predicate.Net)) this.saveService.save.next(this.predicate.Net)
  }

  valid(el): boolean {
    this.error = false
    if (!this.saveService.validCidr(el)) {
      this.error = true
    }
    return !this.error
  }
}
