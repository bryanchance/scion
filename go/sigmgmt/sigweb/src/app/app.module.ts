import { HttpClientModule } from '@angular/common/http'
import { NgModule } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { BrowserModule } from '@angular/platform-browser'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { RouterModule } from '@angular/router'
import { JwtModule } from '@auth0/angular-jwt'
import { MarkdownModule, MarkedOptions } from 'ngx-markdown'

import { environment } from '../environments/environment'
import { ApiModule } from './api/api.module'
import { getToken } from './api/user.service'
import { AppRoutingModule } from './app-routing/app-routing.module'
import { AppComponent, OfflineDialogComponent } from './app.component'
import { AuthGuard } from './auth.guard'
import { ConfigComponent } from './config/config.component'
import { ContactComponent } from './contact/contact.component'
import { LicensesComponent } from './licenses/licenses.component'
import { LoginComponent } from './login/login.component'
import { MaterialModule } from './material/material.module'
import { SitesModule } from './sites/sites.module'
import './utils/swap'

@NgModule({
  declarations: [
    AppComponent,
    ContactComponent,
    LoginComponent,
    ConfigComponent,
    LicensesComponent,
    OfflineDialogComponent,
  ],
  imports: [
    BrowserModule,
    MaterialModule,
    BrowserAnimationsModule,
    ApiModule,
    AppRoutingModule,
    SitesModule,
    FormsModule,
    RouterModule,
    HttpClientModule,
    JwtModule.forRoot({
      config: {
        tokenGetter: getToken,
        whitelistedDomains: [environment.domain],
        skipWhenExpired: true
      }
    }),
    MarkdownModule.forRoot({
      provide: MarkedOptions,
      useValue: {
        baseUrl: 'doc/',
      },
    })
  ],
  providers: [
    AuthGuard
  ],
  entryComponents: [
    OfflineDialogComponent
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
