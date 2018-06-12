import { Component, Input } from '@angular/core'

import {
  CondAllOf,
  CondAnyOf,
  CondClass,
  CondIPv4,
  CondNot,
  MatchDestination,
  MatchDSCP,
  MatchSource,
} from '../../../models/cond'

@Component({
  selector: 'ana-all',
  templateUrl: './all.component.html',
  styleUrls: ['./all.component.scss']
})
export class AllComponent {
  @Input() cond: CondAllOf
  showAdd = false

  insert(idx, condType: string) {
    switch (condType) {
      case 'all': {
        const all = new CondAllOf()
        all.Conds.push(this.cond.Conds[idx])
        this.cond.Conds.splice(idx, 1, all)
        break
      }
      case 'any': {
        const any = new CondAnyOf()
        any.Conds.push(this.cond.Conds[idx])
        this.cond.Conds.splice(idx, 1, any)
        break
      }
      case 'not': {
        const not = new CondNot()
        not.Operand =  this.cond.Conds[idx]
        this.cond.Conds.splice(idx, 1, not)
      }
    }
  }

  addAny() {
    this.cond.Conds.unshift(new CondAnyOf())
    this.closeMenu()
  }

  addAll() {
    this.cond.Conds.unshift(new CondAllOf())
    this.closeMenu()
  }

  addNot() {
    this.cond.Conds.unshift(new CondNot())
    this.closeMenu()
  }

  addSrc() {
    const predicate = new MatchSource()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.cond.Conds.unshift(condIPv4)
    this.closeMenu()
  }

  addDst() {
    const predicate = new MatchDestination()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.cond.Conds.unshift(condIPv4)
    this.closeMenu()
  }

  addDscp() {
    const predicate = new MatchDSCP()
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = predicate
    this.cond.Conds.unshift(condIPv4)
    this.closeMenu()
  }

  addCls() {
    const condCls = new CondClass()
    this.cond.Conds.unshift(condCls)
    this.closeMenu()
  }

  delete(idx) {
    this.cond.Conds.splice(idx, 1)
    this.closeMenu()
  }

  closeMenu() {
    this.showAdd = false
  }
}
