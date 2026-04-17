import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-produtos',
  imports: [FormsModule],
  templateUrl: './produtos.html',
  styleUrl: './produtos.css',
})
export class Produtos {

  private readonly http = inject(HttpClient);

  produto = {
    descricao: '',
    saldo: 0,
    preco: 0
  }

  salvarProduto() {
    const url = 'http://localhost:8080/produtos';

    this.http.post(url, this.produto).subscribe({
      next: (response) => {
        console.log("Sucesso! O Go respondeu:", response);
          alert("Produto salvo com sucesso!");

          // Limpa o formulário para o próximo
          this.produto = { descricao: '', saldo: 0, preco: 0 };
      },
      error: (err) => {
          console.error("Ops! Deu erro na entrega:", err);
        }
    });

  }

}
