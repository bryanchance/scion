import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { SiteDetailsComponent } from './site-details.component'
import { HttpHandler } from '@angular/common/http'
import { TestingModule } from '../../testing/testing.module'
import { SiteConfigurationComponent } from '../site-configuration/site-configuration.component'
import { ASListComponent } from '../as/aslist/aslist.component'

describe('SiteDetailsComponent', () => {
  let component: SiteDetailsComponent
  let fixture: ComponentFixture<SiteDetailsComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ SiteDetailsComponent, SiteConfigurationComponent, ASListComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(SiteDetailsComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
