import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { TrafficClassDetailsComponent } from './traffic-class-details.component';

describe('TrafficClassDetailsComponent', () => {
  let component: TrafficClassDetailsComponent;
  let fixture: ComponentFixture<TrafficClassDetailsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ TrafficClassDetailsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TrafficClassDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
