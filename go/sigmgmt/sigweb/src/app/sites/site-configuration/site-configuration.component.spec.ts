import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { SiteConfigurationComponent } from './site-configuration.component'
import { TestingModule } from '../../testing/testing.module'
import { HttpHandler } from '@angular/common/http'
import { Site } from '../models/models'

describe('SiteConfigurationComponent', () => {
  let component: SiteConfigurationComponent
  let fixture: ComponentFixture<SiteConfigurationComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [SiteConfigurationComponent]
    })
      .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(SiteConfigurationComponent)
    component = fixture.componentInstance
    component.site = new Site
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
