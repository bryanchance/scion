import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { SitesComponent } from './sites.component'
import { TestingModule } from '../testing/testing.module'

describe('SitesComponent', () => {
  let component: SitesComponent
  let fixture: ComponentFixture<SitesComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      declarations: [ SitesComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(SitesComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
