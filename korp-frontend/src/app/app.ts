import { Component, signal } from '@angular/core';
import { Produtos } from './produtos/produtos';
import { NotasFiscais } from './notas-fiscais/notas-fiscais';

@Component({
  selector: 'app-root',
  imports: [Produtos, NotasFiscais],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('korp-frontend');
}
