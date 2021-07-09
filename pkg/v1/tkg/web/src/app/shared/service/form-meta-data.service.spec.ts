import { TestBed } from '@angular/core/testing';

import { FormMetaDataService } from './form-meta-data.service';

describe('FormMetaDataService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: FormMetaDataService = TestBed.get(FormMetaDataService);
    expect(service).toBeTruthy();
  });
});
