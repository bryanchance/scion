import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DscpComponent } from './dscp.component';

describe('DscpComponent', () => {
  let component: DscpComponent;
  let fixture: ComponentFixture<DscpComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DscpComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DscpComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
