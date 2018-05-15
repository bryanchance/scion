import { TestBed, async, inject } from '@angular/core/testing'

import { AuthGuard } from './auth.guard'
import { TestingModule } from './testing/testing.module'
import { HttpHandler } from '@angular/common/http'

describe('AuthGuard', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [AuthGuard, HttpHandler  ]
    })
  })

  it('should ...', inject([AuthGuard], (guard: AuthGuard) => {
    expect(guard).toBeTruthy()
  }))
})
