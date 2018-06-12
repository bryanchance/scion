import { Component, EventEmitter, Input, Output } from '@angular/core'

import { CondIPv4 } from '../../../models/cond'

@Component({
  selector: 'ana-ipv4',
  templateUrl: './ipv4.component.html',
  styleUrls: ['./ipv4.component.scss']
})
export class Ipv4Component {
  @Input() predicate: CondIPv4
}
