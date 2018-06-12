import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TrafficClassesComponent } from './traffic-classes.component';

describe('TrafficClassesComponent', () => {
  let component: TrafficClassesComponent;
  let fixture: ComponentFixture<TrafficClassesComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ TrafficClassesComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TrafficClassesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
