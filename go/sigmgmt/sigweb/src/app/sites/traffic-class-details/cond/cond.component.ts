import { Component, EventEmitter, Input, Output } from '@angular/core'

import {
  Cond,
  TypeCondAllOf,
  TypeCondAnyOf,
  TypeCondClass,
  TypeCondIPv4,
  TypeCondNot,
  TypeIPv4MatchDestination,
  TypeIPv4MatchDSCP,
  TypeIPv4MatchSource,
} from '../../models/cond'

@Component({
  selector: 'ana-cond',
  templateUrl: './cond.component.html',
  styleUrls: ['./cond.component.scss']
})
export class CondComponent {
  @Input() cond: Cond
  @Output() deleted = new EventEmitter()
  @Output() insert = new EventEmitter<string>()
  overDelete: boolean
  showInsert = false

  get anyType(): string {
    return TypeCondAnyOf
  }

  get allType(): string {
    return TypeCondAllOf
  }

  get notType(): string {
    return TypeCondNot
  }

  get ipv4Type(): string {
    return TypeCondIPv4
  }

  get classType(): string {
    return TypeCondClass
  }

  get srcType(): string {
    return TypeIPv4MatchSource
  }

  get dstType(): string {
    return TypeIPv4MatchDestination
  }

  get dscpType(): string {
    return TypeIPv4MatchDSCP
  }

  get showDelete(): boolean {
    return ![this.dscpType, this.dstType, this.srcType].includes(this.cond.Type())
  }
}
