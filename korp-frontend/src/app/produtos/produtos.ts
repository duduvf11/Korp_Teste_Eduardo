import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { HttpClient } from '@angular/common/http';
import { CommonModule, CurrencyPipe } from '@angular/common';

type Produto = {
  codigo: number;
  descricao: string;
  saldo: number;
  preco: number;
};

@Component({
  selector: 'app-produtos',
  imports: [FormsModule, CommonModule, CurrencyPipe],
  templateUrl: './produtos.html',
  styleUrl: './produtos.css',
})
export class Produtos implements OnInit {
  private readonly http = inject(HttpClient);

  produto: Omit<Produto, 'codigo'> = { descricao: '', saldo: 0, preco: 0 };
  listaProdutos = signal<Produto[]>([]);

  ngOnInit() {
    this.carregarProdutos();
  }

  carregarProdutos() {
    const url = 'http://localhost:8080/produtos';

    this.http.get<Produto[]>(url).subscribe({
      next: (response) => {
        this.listaProdutos.set(response);
        console.log('A lista chegou do Go:', this.listaProdutos());
      },
      error: (err) => {
        console.error('Ops! Deu erro ao carregar os produtos:', err);
      }
    });

  }

  salvarProduto() {
    const url = 'http://localhost:8080/produtos';

    this.http.post(url, this.produto).subscribe({
      next: (response) => {
        console.log('Sucesso! O Go respondeu:', response);
        alert('Produto salvo com sucesso!');

        this.produto = { descricao: '', saldo: 0, preco: 0 };
        this.carregarProdutos();
      },
      error: (err) => {
          console.error('Ops! Deu erro na entrega:', err);
        }
    });

  }
}
