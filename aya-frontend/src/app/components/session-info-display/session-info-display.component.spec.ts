import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SessionInfoDisplayComponent } from './session-info-display.component';
import { signal } from '@angular/core';

describe('SessionInfoDisplayComponent', () => {
  let component: SessionInfoDisplayComponent;
  let fixture: ComponentFixture<SessionInfoDisplayComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SessionInfoDisplayComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(SessionInfoDisplayComponent);
    component = fixture.componentInstance;
    component.sessionInfo = signal({
      ID: 1,
      UUID: '',
      Resources: '[]',
      IsOn: true,
      UserID: 2,
    });
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
