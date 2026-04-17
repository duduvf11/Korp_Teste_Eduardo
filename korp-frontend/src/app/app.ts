import { Component, signal } from '@angular/core';
import { Produtos } from './produtos/produtos';

@Component({
  selector: 'app-root',
  imports: [Produtos],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('korp-frontend');
}
