import { Component, EventEmitter, Input, Output } from '@angular/core'

import {
  CondAllOf,
  CondAnyOf,
  CondBool,
  CondClass,
  CondIPv4,
  CondNot,
  MatchDestination,
  MatchDSCP,
  MatchSource,
} from '../../../models/cond'

@Component({
  selector: 'ana-not',
  templateUrl: './not.component.html',
  styleUrls: ['./not.component.scss']
})
export class NotComponent {
  @Input() cond: CondNot
  @Output() over = new EventEmitter<boolean>()
  showAdd = true

  insert(condType: string) {
    switch (condType) {
      case 'all': {
        const all = new CondAllOf()
        all.Conds.push(this.cond.Operand)
        this.cond.Operand = all
        break
      }
      case 'any': {
        const any = new CondAnyOf()
        any.Conds.push(this.cond.Operand)
        this.cond.Operand = any
        break
      }
    }
  }

  addAny() {
    this.cond.Operand = new CondAnyOf()
  }

  addAll() {
    this.cond.Operand = new CondAllOf()
  }

  addBool() {
    this.cond.Operand = new CondBool()
  }

  addSrc() {
    const predicate = new MatchSource()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.cond.Operand = condIPv4
  }

  addDst() {
    const predicate = new MatchDestination()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.cond.Operand = condIPv4
  }

  addDscp() {
    const predicate = new MatchDSCP()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.cond.Operand = condIPv4
  }

  addCls() {
    const condCls = new CondClass()
    this.cond.Operand = condCls
  }

  delete() {
    delete this.cond.Operand
  }
}
