import { Component } from '@angular/core';
import { MatToolbar } from '@angular/material/toolbar';
import { MatIcon } from '@angular/material/icon';
import { MatIconButton } from '@angular/material/button';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [MatToolbar, MatIcon, MatIconButton],
  templateUrl: 'home.page.html',
  styleUrl: 'home.page.css',
})
export default class HomePage {}
