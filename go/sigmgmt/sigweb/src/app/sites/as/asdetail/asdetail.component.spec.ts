import { HttpHandler } from '@angular/common/http'
import { async, ComponentFixture, TestBed } from '@angular/core/testing'
import { ActivatedRoute } from '@angular/router'

import { TestingModule } from '../../../testing/testing.module'
import { ASDetailComponent } from './asdetail.component'
import { ClassifiersComponent } from './classifiers/classifiers.component'
import { NetworksComponent } from './networks/networks.component'
import { SigsComponent } from './sigs/sigs.component'

describe('AsdetailComponent', () => {
  let component: ASDetailComponent
  let fixture: ComponentFixture<ASDetailComponent>

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [TestingModule],
      providers: [HttpHandler,
        {
          provide: ActivatedRoute, useValue: {
            snapshot: {
              params: { ia: '1-1', site: 'site' }
            }
          }
        }],
      declarations: [ASDetailComponent, SigsComponent, NetworksComponent, ClassifiersComponent]
    })
      .compileComponents()
  }))

  beforeEach(() => {
    fixture = TestBed.createComponent(ASDetailComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
