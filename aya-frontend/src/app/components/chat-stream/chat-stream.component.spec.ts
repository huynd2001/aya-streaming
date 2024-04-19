import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChatStreamComponent } from './chat-stream.component';

describe('ChatStreamComponent', () => {
  let component: ChatStreamComponent;
  let fixture: ComponentFixture<ChatStreamComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ChatStreamComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(ChatStreamComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
