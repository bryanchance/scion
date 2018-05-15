import { CommonModule } from '@angular/common'
import { NgModule } from '@angular/core'

import { ApiService } from './api.service'
import { httpInterceptorProviders } from './http-interceptors'
import { UserService } from './user.service'

@NgModule({
  imports: [
    CommonModule,
  ],
  providers: [
    ApiService,
    UserService,
    httpInterceptorProviders
  ]
})
export class ApiModule { }
