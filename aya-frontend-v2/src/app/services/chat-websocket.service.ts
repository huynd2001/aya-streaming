import { Injectable } from '@angular/core';
import { MessageUpdate } from '../interfaces/message';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class ChatWebsocketService {
  constructor() {}

  connect(wsURL: string): Observable<MessageUpdate> {
    return webSocket<MessageUpdate>(wsURL).asObservable();
  }
}
