import { HttpClientModule } from '@angular/common/http'
import { NgModule } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { BrowserModule } from '@angular/platform-browser'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { RouterTestingModule } from '@angular/router/testing'
import { JwtModule } from '@auth0/angular-jwt'

import { ApiModule } from '../api/api.module'
import { ApiService } from '../api/api.service'
import { UserService } from '../api/user.service'
import { MaterialModule } from '../material/material.module'

export function tokenGetter() {
  return ''
}

@NgModule({
  imports: [
    BrowserModule,
    MaterialModule,
    BrowserAnimationsModule,
    JwtModule.forRoot({
      config: {
        tokenGetter: tokenGetter
      }
    }),
    ApiModule,
    FormsModule,
    RouterTestingModule,
    HttpClientModule
  ],
  exports: [
    BrowserModule,
    MaterialModule,
    BrowserAnimationsModule,
    JwtModule,
    ApiModule,
    FormsModule,
    RouterTestingModule,
  ],
  providers: [ApiService, UserService],
})
export class TestingModule { }
