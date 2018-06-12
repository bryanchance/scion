import { Injectable } from '@angular/core'
import { Validator } from 'ip-num/Validator'
import { BehaviorSubject } from 'rxjs'
import { debounceTime, distinctUntilChanged, skip, tap } from 'rxjs/operators'

import { Site } from '../models/models'

@Injectable({
  providedIn: 'root'
})
export class TrafficClassService {
  save = new BehaviorSubject({})
  site = new BehaviorSubject(new Site(0))
  tcID = 0

  get saver() {
    return this.save.pipe(
      skip(1),
      debounceTime(200),
      distinctUntilChanged()
    )
  }

  validCidr(cidr): boolean {
    return Validator.isValidIPv4CidrNotation(cidr)[0]
  }

  validDSCP(dscp: string): boolean {
    if (dscp.substring(0, 2) !== '0x') return false
    const dVal = parseInt(dscp.substring(2), 16)
    return dVal <= 64 && dVal >= 0
  }
}
