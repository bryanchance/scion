import { HttpClient, HttpHandler } from '@angular/common/http'
import { inject, TestBed } from '@angular/core/testing'
import { JwtHelperService } from '@auth0/angular-jwt'

import { TestingModule } from '../testing/testing.module'
import { ApiService } from './api.service'
import { UserService } from './user.service'

describe('UserService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [UserService, ApiService, JwtHelperService, HttpClient, HttpHandler]
    })
  })

  it('should be created', inject([UserService], (service: UserService) => {
    expect(service).toBeTruthy()
  }))
})
