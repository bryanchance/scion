import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { PathSelectorComponent } from './path-selector.component'
import { TestingModule } from '../../testing/testing.module'
import { Site } from '../models'

describe('PathSelectorComponent', () => {
  let component: PathSelectorComponent
  let fixture: ComponentFixture<PathSelectorComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      declarations: [ PathSelectorComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(PathSelectorComponent)
    component = fixture.componentInstance
    component.site = new Site
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
