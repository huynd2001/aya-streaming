import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SessionInfoDisplayComponent } from './session-info-display.component';

describe('SessionInfoDisplayComponent', () => {
  let component: SessionInfoDisplayComponent;
  let fixture: ComponentFixture<SessionInfoDisplayComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SessionInfoDisplayComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(SessionInfoDisplayComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
