import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { NetworksComponent } from './networks.component'
import { TestingModule } from '../../../../testing/testing.module'
import { HttpHandler } from '@angular/common/http'

describe('NetworksComponent', () => {
  let component: NetworksComponent
  let fixture: ComponentFixture<NetworksComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler],
      declarations: [ NetworksComponent ]
    })
    .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(NetworksComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
