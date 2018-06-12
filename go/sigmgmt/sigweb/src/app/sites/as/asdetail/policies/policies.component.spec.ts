import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { PoliciesComponent } from './policies.component'
import { TestingModule } from '../../../../testing/testing.module'
import { HttpHandler } from '@angular/common/http'

describe('PoliciesComponent', () => {
  let component: PoliciesComponent
  let fixture: ComponentFixture<PoliciesComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ PoliciesComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(PoliciesComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
