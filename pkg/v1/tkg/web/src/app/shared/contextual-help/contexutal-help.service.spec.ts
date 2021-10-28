import { TestBed } from '@angular/core/testing';

import { ContexutalHelpService } from './contexutal-help.service';

describe('ContexutalHelpService', () => {
  let service: ContexutalHelpService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ContexutalHelpService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
