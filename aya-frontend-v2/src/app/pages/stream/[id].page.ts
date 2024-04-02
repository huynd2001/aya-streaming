import { Component, Input } from '@angular/core';
import { ChatStreamComponent } from '../../components/chat-stream/chat-stream.component';

@Component({
  selector: 'app-stream',
  standalone: true,
  imports: [ChatStreamComponent],
  templateUrl: '[id].page.html',
})
export default class IdPage {
  @Input() id: string | undefined;
}
