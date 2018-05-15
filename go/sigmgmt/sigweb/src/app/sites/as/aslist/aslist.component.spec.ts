import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { ASListComponent } from './aslist.component'
import { HttpHandler } from '@angular/common/http'
import { TestingModule } from '../../../testing/testing.module'

describe('AslistComponent', () => {
  let component: ASListComponent
  let fixture: ComponentFixture<ASListComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ ASListComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(ASListComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
