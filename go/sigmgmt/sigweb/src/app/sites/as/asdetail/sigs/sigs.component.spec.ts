import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { SigsComponent } from './sigs.component'
import { TestingModule } from '../../../../testing/testing.module'
import { HttpHandler } from '@angular/common/http'

describe('SigsComponent', () => {
  let component: SigsComponent
  let fixture: ComponentFixture<SigsComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ SigsComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(SigsComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
