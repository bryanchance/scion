import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { Ipv4Component } from './ipv4.component';

describe('Ipv4Component', () => {
  let component: Ipv4Component;
  let fixture: ComponentFixture<Ipv4Component>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ Ipv4Component ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(Ipv4Component);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
