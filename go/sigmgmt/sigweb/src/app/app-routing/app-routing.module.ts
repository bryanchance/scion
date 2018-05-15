import { CommonModule } from '@angular/common'
import { NgModule } from '@angular/core'
import { RouterModule, Routes } from '@angular/router'

import { AuthGuard } from '../auth.guard'
import { ConfigComponent } from '../config/config.component'
import { ContactComponent } from '../contact/contact.component'
import { LoginComponent } from '../login/login.component'
import { ASDetailComponent } from '../sites/as/asdetail/asdetail.component'
import { SiteDetailsComponent } from '../sites/site-details/site-details.component'
import { SitesComponent } from '../sites/sites.component'

const appRoutes: Routes = [
  { path: '', redirectTo: '/sites', pathMatch: 'full' },
  { path: 'sites', component: SitesComponent, canActivate: [AuthGuard] },
  { path: 'sites/new', component: SiteDetailsComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site', component: SiteDetailsComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/ias/:ia', component: ASDetailComponent, canActivate: [AuthGuard] },
  { path: 'config', component: ConfigComponent },
  { path: 'contact', component: ContactComponent },
  { path: 'login', component: LoginComponent },
]

@NgModule({
  imports: [
    CommonModule,
    RouterModule.forRoot(
      appRoutes,
      // { enableTracing: true } // <-- debugging purposes only
    )
  ],
  exports: [
    RouterModule
  ]
})
export class AppRoutingModule { }
