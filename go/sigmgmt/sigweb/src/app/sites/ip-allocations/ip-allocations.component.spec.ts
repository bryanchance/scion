import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { IpAllocationsComponent } from './ip-allocations.component';

describe('IpAllocationsComponent', () => {
  let component: IpAllocationsComponent;
  let fixture: ComponentFixture<IpAllocationsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ IpAllocationsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(IpAllocationsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
