import { TestBed } from '@angular/core/testing';

import { UserInfoService } from './user-info.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('UserInfoService', () => {
  let service: UserInfoService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(UserInfoService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
