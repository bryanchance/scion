import { CommonModule } from '@angular/common'
import { NgModule } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { RouterModule } from '@angular/router'

import { MaterialModule } from '../material/material.module'
import { ASDetailComponent } from './as/asdetail/asdetail.component'
import { ClassifiersComponent } from './as/asdetail/classifiers/classifiers.component'
import { NetworksComponent } from './as/asdetail/networks/networks.component'
import { SigsComponent } from './as/asdetail/sigs/sigs.component'
import { ASListComponent } from './as/aslist/aslist.component'
import { PathSelectorComponent } from './path-selector/path-selector.component'
import { SiteConfigurationComponent } from './site-configuration/site-configuration.component'
import { SiteDetailsComponent } from './site-details/site-details.component'
import { SitesComponent } from './sites.component'

@NgModule({
  imports: [
    CommonModule,
    MaterialModule,
    RouterModule,
    FormsModule
  ],
  declarations: [
    SitesComponent,
    SiteDetailsComponent,
    SiteConfigurationComponent,
    PathSelectorComponent,
    ASListComponent,
    ASDetailComponent,
    NetworksComponent,
    SigsComponent,
    ClassifiersComponent
  ]
})
export class SitesModule { }
