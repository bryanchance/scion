import { CommonModule } from '@angular/common'
import { NgModule } from '@angular/core'
import { RouterModule, Routes } from '@angular/router'

import { AuthGuard } from '../auth.guard'
import { ConfigComponent } from '../config/config.component'
import { ContactComponent } from '../contact/contact.component'
import { LicensesComponent } from '../licenses/licenses.component'
import { LoginComponent } from '../login/login.component'
import { ASDetailComponent } from '../sites/as/asdetail/asdetail.component'
import { ASListComponent } from '../sites/as/aslist/aslist.component'
import { PathSelectorComponent } from '../sites/path-selector/path-selector.component'
import { SiteDetailsComponent } from '../sites/site-details/site-details.component'
import { SitesComponent } from '../sites/sites.component'
import { TrafficClassesComponent } from '../sites/traffic-classes/traffic-classes.component'
import { TrafficClassDetailsComponent } from '../sites/traffic-class-details/traffic-class-details.component';

const appRoutes: Routes = [
  { path: '', redirectTo: '/sites', pathMatch: 'full' },
  { path: 'sites', component: SitesComponent, canActivate: [AuthGuard] },
  { path: 'sites/new', component: SiteDetailsComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site', component: SiteDetailsComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/ases', component: ASListComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/ases/:ia', component: ASDetailComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/selectors', component: PathSelectorComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/classes', component: TrafficClassesComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/classes/new', component: TrafficClassDetailsComponent, canActivate: [AuthGuard] },
  { path: 'sites/:site/classes/:class', component: TrafficClassDetailsComponent, canActivate: [AuthGuard] },
  { path: 'config', component: ConfigComponent },
  { path: 'contact', component: ContactComponent },
  { path: 'login', component: LoginComponent },
  { path: 'licenses', component: LicensesComponent },
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
