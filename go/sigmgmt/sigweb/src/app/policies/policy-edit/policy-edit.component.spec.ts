import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { PolicyEditComponent } from './policy-edit.component';

describe('PolicyEditComponent', () => {
  let component: PolicyEditComponent;
  let fixture: ComponentFixture<PolicyEditComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PolicyEditComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PolicyEditComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
