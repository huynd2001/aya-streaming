import { Component, Input } from '@angular/core';
import { ChatStreamComponent } from '../../components/chat-stream/chat-stream.component';

@Component({
  selector: 'app-stream',
  standalone: true,
  imports: [ChatStreamComponent],
  template: `<app-chat-stream streamId="{{ id }}"></app-chat-stream>`,
})
export default class IdPage {
  @Input() id: string | undefined;
}
