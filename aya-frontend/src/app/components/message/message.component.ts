import { Component, Input, OnInit } from '@angular/core';
import {
  animate,
  state,
  style,
  transition,
  trigger,
} from '@angular/animations';
import { DisplayMessage } from '../../interfaces/message';
import { NgForOf, NgIf, NgOptimizedImage } from '@angular/common';

@Component({
  selector: 'app-message',
  standalone: true,
  imports: [NgForOf, NgIf, NgOptimizedImage],
  templateUrl: 'message.component.html',
  styleUrl: 'message.component.css',
  animations: [
    trigger('push', [
      state(
        'init',
        style({
          height: '0px',
          opacity: 0,
        })
      ),
      state(
        'loaded',
        style({
          opacity: 1,
        })
      ),
      state(
        'removed',
        style({
          opacity: 0,
        })
      ),
      state(
        'deleted',
        style({
          height: '0px',
          opacity: 0,
        })
      ),
      transition('init => loaded', [animate('0.3s')]),
      transition('loaded => removed', [animate('0.3s')]),
      transition('loaded => deleted', [animate('0.3s 1s')]),
    ]),
  ],
})
export class MessageComponent implements OnInit {
  @Input() displayMsg: DisplayMessage | undefined;
  ngOnInit(): void {
    // IDK why but I put this in my legacy code
  }

  getIcon(): string {
    return this.displayMsg?.message.author.isBot ||
      this.displayMsg?.message.author.isAdmin
      ? this.displayMsg.message.author.isBot
        ? 'bi bi-gear'
        : 'bi bi-shield-fill-check'
      : '';
  }

  getSource(): string {
    switch (this.displayMsg?.message.source) {
      case 'discord':
        return '/discord.svg';
      case 'youtube':
        return '/youtube.svg';
      case 'twitch':
        return '/twitch.svg';
      default:
        return '/analog.svg';
    }
  }

  getState(): string {
    if (this.displayMsg?.init) return 'init';
    else if (this.displayMsg?.delete) return 'deleted';
    else if (this.displayMsg?.remove) return 'removed';
    else return 'loaded';
  }
}
