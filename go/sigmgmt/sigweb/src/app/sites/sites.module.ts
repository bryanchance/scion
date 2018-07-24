import { CommonModule } from '@angular/common'
import { NgModule } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { RouterModule } from '@angular/router'
import { NgSelectModule } from '@ng-select/ng-select'

import { MaterialModule } from '../material/material.module'
import { MouseoverDirective } from '../mouseover.directive'
import { ASDetailComponent } from './as/asdetail/asdetail.component'
import { PoliciesComponent } from './as/asdetail/policies/policies.component'
import { NetworksComponent } from './as/asdetail/networks/networks.component'
import { SigsComponent } from './as/asdetail/sigs/sigs.component'
import { ASListComponent } from './as/aslist/aslist.component'
import { PathSelectorComponent } from './path-selector/path-selector.component'
import { SiteConfigurationComponent } from './site-configuration/site-configuration.component'
import { SiteDetailsComponent } from './site-details/site-details.component'
import { SitesComponent } from './sites.component'
import { AllComponent } from './traffic-class-details/cond/all/all.component'
import { AnyComponent } from './traffic-class-details/cond/any/any.component'
import { CondComponent } from './traffic-class-details/cond/cond.component'
import { DscpComponent } from './traffic-class-details/cond/dscp/dscp.component'
import { DstComponent } from './traffic-class-details/cond/dst/dst.component'
import { Ipv4Component } from './traffic-class-details/cond/ipv4/ipv4.component'
import { NotComponent } from './traffic-class-details/cond/not/not.component'
import { SrcComponent } from './traffic-class-details/cond/src/src.component'
import { TrafficClassDetailsComponent } from './traffic-class-details/traffic-class-details.component'
import { TrafficClassesComponent } from './traffic-classes/traffic-classes.component';
import { ClassComponent } from './traffic-class-details/cond/class/class.component';
import { BoolComponent } from './traffic-class-details/cond/bool/bool.component'

@NgModule({
  imports: [
    CommonModule,
    MaterialModule,
    RouterModule,
    FormsModule,
    NgSelectModule
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
    PoliciesComponent,
    TrafficClassesComponent,
    CondComponent,
    NotComponent,
    AllComponent,
    AnyComponent,
    SrcComponent,
    DstComponent,
    DscpComponent,
    Ipv4Component,
    TrafficClassDetailsComponent,
    MouseoverDirective,
    ClassComponent,
    BoolComponent,
  ],
})
export class SitesModule { }
