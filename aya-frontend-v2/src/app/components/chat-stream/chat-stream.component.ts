import { Component, Input, OnDestroy, OnInit } from '@angular/core';
import { NgForOf, NgIf } from '@angular/common';
import { MessageComponent } from '../message/message.component';
import { ChatWebsocketService } from '../../services/chat-websocket.service';
import { Observable, Subscription } from 'rxjs';
import {
  DisplayMessage,
  Message,
  MessageUpdate,
} from '../../interfaces/message';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-chat-stream',
  standalone: true,
  imports: [NgIf, NgForOf, MessageComponent],
  templateUrl: 'chat-stream.component.html',
  styleUrl: 'chat-stream.component.css',
})
export class ChatStreamComponent implements OnInit, OnDestroy {
  @Input() streamId: string = 'undefined';

  width: number = 400;
  height: number = 600;
  isConnected: boolean = false;
  maxMessages: number = 100;
  timeOut: number = 60000;
  private chatSubscription: Subscription | undefined;
  displayMessages: DisplayMessage[] = [];

  getColor(r: number, g: number, b: number): string {
    return `rgba(${r}, ${g}, ${b}, 100)`;
  }

  constructor(private chatWebsocketService: ChatWebsocketService) {}

  ngOnInit(): void {
    let url = `${environment.wsSocketUrl}/stream/${this.streamId}`;
    let chatObs = this.chatWebsocketService.connect(url);
    this.isConnected = true;
    this.chatSubscription = chatObs.subscribe({
      next: (messageUpdate: MessageUpdate) => {
        console.log(messageUpdate);
        this.updateMessage(messageUpdate.message, messageUpdate.update);
      },
      error: (error) => console.error(error),
    });
  }

  updateMessage(msg: Message, update: 'new' | 'delete' | 'edit') {
    switch (update) {
      case 'new': {
        let displayMsg: DisplayMessage = {
          message: msg,
          init: true,
          remove: false,
          delete: false,
        };
        this.displayMessages.push(displayMsg);
        if (this.displayMessages.find.length > this.maxMessages) {
          this.displayMessages.shift();
        }
        new Promise((r) => setTimeout(r, 5)).then(() => {
          displayMsg.init = false;
        });
        new Promise((r) => setTimeout(r, this.timeOut))
          .then(() => {
            displayMsg.remove = true;
          })
          .then(() => {
            new Promise((r) => setTimeout(r, 500)).then(() => {
              if (this.displayMessages.find((m) => m.message.id == msg.id))
                this.displayMessages.shift();
            });
          });
        break;
      }
      case 'edit': {
        let displayMsg = this.displayMessages.find(
          (dMsg) => dMsg.message.id == msg.id
        );
        if (displayMsg) {
          displayMsg.message = msg;
          displayMsg.edit = true;
        }
        break;
      }
      case 'delete': {
        let displayMsg = this.displayMessages.find(
          (dMsg) => dMsg.message.id == msg.id
        );
        if (displayMsg) {
          displayMsg.delete = true;
        }
        break;
      }
    }
  }

  ngOnDestroy(): void {
    if (this.chatSubscription) {
      this.chatSubscription.unsubscribe();
    }
  }
}
