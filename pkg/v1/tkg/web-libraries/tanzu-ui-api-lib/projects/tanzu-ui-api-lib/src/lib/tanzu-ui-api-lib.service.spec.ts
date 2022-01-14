import { TestBed } from '@angular/core/testing';

import { TanzuUiApiLibService } from './tanzu-ui-api-lib.service';

describe('TanzuUiApiLibService', () => {
  let service: TanzuUiApiLibService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(TanzuUiApiLibService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
