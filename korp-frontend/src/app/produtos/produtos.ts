import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { HttpClient } from '@angular/common/http';
import { CommonModule, CurrencyPipe } from '@angular/common';
import { environment } from '../../environments/environment';

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
  private readonly urlProdutos = `${environment.api.estoqueBaseUrl}/produtos`;

  produto: Produto = this.novoProdutoVazio();
  listaProdutos = signal<Produto[]>([]);
  modoEdicao = signal(false);
  codigoProdutoEmEdicao = signal<number | null>(null);
  confirmacaoExclusaoProduto = signal<Produto | null>(null);
  produtoEmExclusaoCodigo = signal<number | null>(null);
  carregando = signal(false);
  mensagemSucesso = signal('');
  mensagemErro = signal('');

  ngOnInit(): void {
    this.carregarProdutos();
  }

  carregarProdutos(): void {
    this.carregando.set(true);

    this.http.get<Produto[]>(this.urlProdutos).subscribe({
      next: (response) => {
        this.listaProdutos.set(this.ordenarProdutosPorCodigo(response ?? []));
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel carregar os produtos.'));
      },
      complete: () => {
        this.carregando.set(false);
      },
    });
  }

  iniciarEdicao(produto: Produto): void {
    this.limparMensagens();
    this.produto = {
      ...produto,
    };
    this.modoEdicao.set(true);
    this.codigoProdutoEmEdicao.set(produto.codigo);
    this.mensagemSucesso.set(`Editando produto de codigo ${produto.codigo}.`);
  }

  cancelarEdicao(): void {
    this.limparMensagens();
    this.resetarFormulario();
    this.mensagemSucesso.set('Edicao cancelada.');
  }

  salvarProduto(): void {
    this.limparMensagens();
    const descricao = this.produto.descricao.trim();

    if (this.produto.codigo <= 0) {
      this.mensagemErro.set('Informe um codigo de produto maior que zero.');
      return;
    }

    if (!descricao) {
      this.mensagemErro.set('Informe a descricao do produto.');
      return;
    }

    if (this.produto.saldo < 0) {
      this.mensagemErro.set('O saldo nao pode ser negativo.');
      return;
    }

    if (this.produto.preco < 0) {
      this.mensagemErro.set('O preco nao pode ser negativo.');
      return;
    }

    const payload: Produto = {
      ...this.produto,
      descricao,
    };

    if (this.modoEdicao() && this.codigoProdutoEmEdicao() !== null) {
      const codigoAtual = this.codigoProdutoEmEdicao() as number;

      this.http.put<Produto>(`${this.urlProdutos}/${codigoAtual}`, payload).subscribe({
        next: () => {
          this.concluirSalvamento('Produto atualizado com sucesso!');
        },
        error: (err) => {
          this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel atualizar o produto.'));
        },
      });

      return;
    }

    this.http.post<Produto>(this.urlProdutos, payload).subscribe({
      next: () => {
        this.concluirSalvamento('Produto salvo com sucesso!');
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel salvar o produto.'));
      },
    });
  }

  solicitarExclusaoProduto(produto: Produto): void {
    this.limparMensagens();
    this.confirmacaoExclusaoProduto.set(produto);
  }

  cancelarExclusaoProduto(): void {
    if (this.produtoEmExclusaoCodigo() !== null) {
      return;
    }

    this.confirmacaoExclusaoProduto.set(null);
  }

  confirmarExclusaoProduto(): void {
    const produto = this.confirmacaoExclusaoProduto();
    if (!produto) {
      return;
    }

    this.produtoEmExclusaoCodigo.set(produto.codigo);

    this.http.delete(`${this.urlProdutos}/${produto.codigo}`).subscribe({
      next: () => {
        if (this.codigoProdutoEmEdicao() === produto.codigo) {
          this.resetarFormulario();
        }

        this.confirmacaoExclusaoProduto.set(null);
        this.mensagemSucesso.set('Produto deletado com sucesso!');
        this.carregarProdutos();
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel deletar o produto.'));
      },
      complete: () => {
        this.produtoEmExclusaoCodigo.set(null);
      },
    });
  }

  textoBotaoConfirmarExclusaoProduto(): string {
    return this.produtoEmExclusaoCodigo() === null ? 'Excluir produto' : 'Excluindo...';
  }

  private concluirSalvamento(mensagem: string): void {
    this.mensagemSucesso.set(mensagem);
    this.resetarFormulario();
    this.carregarProdutos();
  }

  private limparMensagens(): void {
    this.mensagemErro.set('');
    this.mensagemSucesso.set('');
  }

  private resetarFormulario(): void {
    this.produto = this.novoProdutoVazio();
    this.modoEdicao.set(false);
    this.codigoProdutoEmEdicao.set(null);
  }

  private novoProdutoVazio(): Produto {
    return { codigo: 0, descricao: '', saldo: 0, preco: 0 };
  }

  private ordenarProdutosPorCodigo(produtos: Produto[]): Produto[] {
    return [...produtos].sort((a, b) => a.codigo - b.codigo);
  }

  private extrairMensagemErro(err: unknown, fallback: string): string {
    const erroHttp = err as { error?: unknown };
    const resposta = erroHttp?.error;

    if (typeof resposta === 'string' && resposta.trim() !== '') {
      return resposta;
    }

    if (resposta && typeof resposta === 'object') {
      const respostaObj = resposta as Record<string, unknown>;
      const mensagem = respostaObj['erro'];

      if (typeof mensagem === 'string' && mensagem.trim() !== '') {
        return mensagem;
      }
    }

    return fallback;
  }
}
