import { TestBed } from '@angular/core/testing';

import { ContextualHelpService } from './contextual-help.service';

describe('ContexutalHelpService', () => {
  let service: ContextualHelpService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ContextualHelpService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
