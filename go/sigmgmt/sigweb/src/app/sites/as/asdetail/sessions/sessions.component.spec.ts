import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { SessionsComponent } from './sessions.component'
import { TestingModule } from '../../../../testing/testing.module'
import { HttpHandler } from '@angular/common/http'

describe('SessionsComponent', () => {
  let component: SessionsComponent
  let fixture: ComponentFixture<SessionsComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ SessionsComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(SessionsComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
