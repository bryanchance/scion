import { inject, TestBed } from '@angular/core/testing'

import { TrafficClassService } from './traffic-class.service'

describe('SaveService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [TrafficClassService]
    });
  });

  it('should be created', inject([TrafficClassService], (service: TrafficClassService) => {
    expect(service).toBeTruthy();
  }));
});
