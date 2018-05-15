import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { ClassifiersComponent } from './classifiers.component'
import { TestingModule } from '../../../../testing/testing.module'
import { HttpHandler } from '@angular/common/http'

describe('ClassifiersComponent', () => {
  let component: ClassifiersComponent
  let fixture: ComponentFixture<ClassifiersComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ ClassifiersComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(ClassifiersComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
