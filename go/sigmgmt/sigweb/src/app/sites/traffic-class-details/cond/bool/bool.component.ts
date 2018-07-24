import { AfterViewInit, Component, Input, OnDestroy, OnInit, ViewChild } from '@angular/core'
import { NgForm } from '@angular/forms'
import { Subscription } from 'rxjs'
import { auditTime, debounceTime, distinctUntilChanged, map, skip } from 'rxjs/operators'

import { CondBool } from '../../../models/cond'
import { TrafficClassService } from '../../traffic-class.service'

@Component({
  selector: 'ana-bool',
  templateUrl: './bool.component.html',
  styleUrls: ['./bool.component.scss']
})
export class BoolComponent implements OnDestroy, AfterViewInit {
  @Input() cond: CondBool
  @ViewChild('boolForm') form: NgForm
  sub: Subscription
  error = false

  constructor(public saveService: TrafficClassService) { }

  ngAfterViewInit() {
    this.sub = this.form.form.valueChanges.pipe(
      skip(1),
      map(el => el.name),
      debounceTime(500),
      auditTime(400),
      distinctUntilChanged(),
      debounceTime(1000),
    ).subscribe(
      () => this.saveService.save.next(this.cond.Value)
    )
  }

  ngOnDestroy() {
    this.sub.unsubscribe()
  }
}
