import { TestBed } from '@angular/core/testing';

import { SessionInfoService } from './session-info.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('SessionInfoService', () => {
  let service: SessionInfoService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(SessionInfoService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
