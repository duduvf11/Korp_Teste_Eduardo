import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';

interface ItemNota {
  produtoId: number;
  quantidade: number;
}

interface NotaFiscal {
  id: number;
  cliente: string;
  estaAberta: boolean;
  itens: ItemNota[];
};


@Component({
  selector: 'app-notas-fiscais',
  imports: [FormsModule, CommonModule],
  templateUrl: './notas-fiscais.html',
  styleUrl: './notas-fiscais.css',
})
export class NotasFiscais implements OnInit {
  private readonly http = inject(HttpClient);

  notaAtual: NotaFiscal = {
    id: 0,
    cliente: '',
    estaAberta: true,
    itens: []
  };

  produtosDisponiveis: any[] = [];

  ngOnInit() {
    this.carregarProdutos();
  }

  carregarProdutos() {
    const url = 'http://localhost:8080/produtos';
    this.http.get(url).subscribe({
      next: (response: any) => {
        this.produtosDisponiveis = [...response];
      },
      error: (err) => {
        console.error("Erro ao carregar produtos:", err);
      }
    });
  }

}
