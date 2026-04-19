import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CommonModule, CurrencyPipe } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

type Produto = {
  codigo: number;
  descricao: string;
  saldo: number;
  preco: number;
};

type ItemNotaPayload = {
  produto_id: number;
  quantidade: number;
};

type ItemNota = {
  id: number;
  produto_id: number;
  nota_fiscal_id: number;
  quantidade: number;
};

type NotaFiscal = {
  id: number;
  cliente: string;
  esta_aberta: boolean;
  itens: ItemNota[];
};

type RespostaImpressaoNota = {
  mensagem: string;
  nota: NotaFiscal;
};

type RespostaCancelamentoNota = {
  mensagem: string;
  nota: NotaFiscal;
};

type RespostaExclusaoNota = {
  mensagem: string;
  nota_id: number;
};

@Component({
  selector: 'app-notas-fiscais',
  imports: [FormsModule, CommonModule, CurrencyPipe],
  templateUrl: './notas-fiscais.html',
  styleUrl: './notas-fiscais.css',
})
export class NotasFiscais implements OnInit {
  private readonly http = inject(HttpClient);

  private readonly urlProdutos = `${environment.api.estoqueBaseUrl}/produtos`;
  private readonly urlNotasFiscais = `${environment.api.faturamentoBaseUrl}/notas-fiscais`;

  produtosDisponiveis = signal<Produto[]>([]);
  itensNotaAtual = signal<ItemNotaPayload[]>([]);
  notasFiscais = signal<NotaFiscal[]>([]);

  clienteNotaAtual = '';
  produtoSelecionadoCodigo: number | null = null;
  quantidadeSelecionada = 1;

  carregandoProdutos = signal(false);
  carregandoNotas = signal(false);
  salvandoNota = signal(false);
  notaEmImpressaoId = signal<number | null>(null);
  notaEmCancelamentoId = signal<number | null>(null);
  notaEmExclusaoId = signal<number | null>(null);
  confirmacaoExclusaoNota = signal<NotaFiscal | null>(null);

  mensagemSucesso = signal('');
  mensagemErro = signal('');

  ngOnInit(): void {
    this.carregarProdutos();
    this.carregarNotas();
  }

  carregarProdutos(): void {
    this.carregandoProdutos.set(true);

    this.http.get<Produto[]>(this.urlProdutos).subscribe({
      next: (response) => {
        const produtosValidos = this.normalizarProdutos(response ?? []);
        this.produtosDisponiveis.set(produtosValidos);

        if (produtosValidos.length === 0) {
          this.mensagemErro.set('Nenhum produto com codigo valido foi encontrado. Cadastre produtos com codigo maior que zero.');
        }
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel carregar os produtos.'));
      },
      complete: () => {
        this.carregandoProdutos.set(false);
      },
    });
  }

  carregarNotas(): void {
    this.carregandoNotas.set(true);

    this.http.get<NotaFiscal[]>(this.urlNotasFiscais).subscribe({
      next: (response) => {
        const notasNormalizadas = (response ?? []).map((nota) => ({
          ...nota,
          itens: nota.itens ?? [],
        }));

        this.notasFiscais.set(this.ordenarNotasPorStatus(notasNormalizadas));
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel carregar as notas fiscais.'));
      },
      complete: () => {
        this.carregandoNotas.set(false);
      },
    });
  }

  adicionarItem(): void {
    this.limparMensagens();

    if (this.produtoSelecionadoCodigo === null) {
      this.mensagemErro.set('Selecione um produto para adicionar na nota.');
      return;
    }

    if (this.quantidadeSelecionada <= 0) {
      this.mensagemErro.set('A quantidade deve ser maior que zero.');
      return;
    }

    const produto = this.produtosDisponiveis().find((item) => item.codigo === this.produtoSelecionadoCodigo);
    const quantidadeJaSelecionada = this.itensNotaAtual()
      .filter((item) => item.produto_id === this.produtoSelecionadoCodigo)
      .reduce((total, item) => total + item.quantidade, 0);

    const novaQuantidadeTotal = quantidadeJaSelecionada + this.quantidadeSelecionada;
    if (produto && novaQuantidadeTotal > produto.saldo) {
      this.mensagemErro.set(
        `Quantidade maior que o saldo disponivel para ${produto.descricao}. Saldo atual: ${produto.saldo}.`,
      );
      return;
    }

    this.itensNotaAtual.update((itens) => {
      const indiceExistente = itens.findIndex((item) => item.produto_id === this.produtoSelecionadoCodigo);

      if (indiceExistente === -1) {
        return [
          ...itens,
          {
            produto_id: this.produtoSelecionadoCodigo as number,
            quantidade: this.quantidadeSelecionada,
          },
        ];
      }

      const copiaItens = [...itens];
      copiaItens[indiceExistente] = {
        ...copiaItens[indiceExistente],
        quantidade: copiaItens[indiceExistente].quantidade + this.quantidadeSelecionada,
      };

      return copiaItens;
    });

    this.mensagemSucesso.set('Item adicionado na nota fiscal.');
    this.produtoSelecionadoCodigo = null;
    this.quantidadeSelecionada = 1;
  }

  removerItem(produtoCodigo: number): void {
    this.limparMensagens();
    this.itensNotaAtual.update((itens) => itens.filter((item) => item.produto_id !== produtoCodigo));
    this.mensagemSucesso.set('Item removido da nota fiscal.');
  }

  salvarNotaFiscal(): void {
    this.limparMensagens();

    const cliente = this.clienteNotaAtual.trim();
    const itens = this.itensNotaAtual();

    if (!cliente) {
      this.mensagemErro.set('Informe o nome do cliente antes de salvar a nota.');
      return;
    }

    if (itens.length === 0) {
      this.mensagemErro.set('Adicione ao menos um item antes de salvar a nota.');
      return;
    }

    this.salvandoNota.set(true);

    this.http
      .post<NotaFiscal>(this.urlNotasFiscais, {
        cliente,
        itens,
      })
      .subscribe({
        next: () => {
          this.mensagemSucesso.set('Nota fiscal criada com sucesso. Status inicial: Aberta.');
          this.clienteNotaAtual = '';
          this.itensNotaAtual.set([]);
          this.produtoSelecionadoCodigo = null;
          this.quantidadeSelecionada = 1;
          this.carregarNotas();
        },
        error: (err) => {
          this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel criar a nota fiscal.'));
        },
        complete: () => {
          this.salvandoNota.set(false);
        },
      });
  }

  imprimirNota(nota: NotaFiscal): void {
    this.limparMensagens();

    if (!nota.esta_aberta) {
      this.mensagemErro.set('Somente notas com status Aberta podem ser impressas.');
      return;
    }

    this.notaEmImpressaoId.set(nota.id);

    this.http.post<RespostaImpressaoNota>(`${this.urlNotasFiscais}/${nota.id}/imprimir`, {}).subscribe({
      next: (response) => {
        const notaAtualizada = response?.nota;

        if (notaAtualizada) {
          this.notasFiscais.update((notas) =>
            this.ordenarNotasPorStatus(
              notas.map((item) =>
                item.id === nota.id
                  ? {
                      ...notaAtualizada,
                      itens: notaAtualizada.itens ?? [],
                    }
                  : item,
              ),
            ),
          );
        }

        this.mensagemSucesso.set(response?.mensagem ?? 'Nota fiscal impressa com sucesso.');
        this.carregarNotas();
        this.carregarProdutos();
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel imprimir a nota fiscal.'));
      },
      complete: () => {
        this.notaEmImpressaoId.set(null);
      },
    });
  }

  cancelarNota(nota: NotaFiscal): void {
    this.limparMensagens();

    if (!nota.esta_aberta) {
      this.mensagemErro.set('Apenas notas em aberto podem ser canceladas.');
      return;
    }

    this.notaEmCancelamentoId.set(nota.id);

    this.http.post<RespostaCancelamentoNota>(`${this.urlNotasFiscais}/${nota.id}/cancelar`, {}).subscribe({
      next: (response) => {
        const notaAtualizada = response?.nota;

        if (notaAtualizada) {
          this.notasFiscais.update((notas) =>
            this.ordenarNotasPorStatus(
              notas.map((item) =>
                item.id === nota.id
                  ? {
                      ...notaAtualizada,
                      itens: notaAtualizada.itens ?? [],
                    }
                  : item,
              ),
            ),
          );
        }

        this.mensagemSucesso.set(response?.mensagem ?? 'Nota fiscal cancelada com sucesso.');
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel cancelar a nota fiscal.'));
      },
      complete: () => {
        this.notaEmCancelamentoId.set(null);
      },
    });
  }

  solicitarExclusaoNota(nota: NotaFiscal): void {
    this.limparMensagens();

    if (nota.esta_aberta) {
      this.mensagemErro.set('Apenas notas fechadas podem ser deletadas.');
      return;
    }

    if (this.estaProcessandoNota(nota.id)) {
      return;
    }

    this.confirmacaoExclusaoNota.set(nota);
  }

  cancelarExclusaoNota(): void {
    if (this.notaEmExclusaoId() !== null) {
      return;
    }

    this.confirmacaoExclusaoNota.set(null);
  }

  confirmarExclusaoNota(): void {
    const nota = this.confirmacaoExclusaoNota();
    if (!nota) {
      return;
    }

    this.notaEmExclusaoId.set(nota.id);

    this.http.delete<RespostaExclusaoNota>(`${this.urlNotasFiscais}/${nota.id}`).subscribe({
      next: (response) => {
        this.notasFiscais.update((notas) => this.ordenarNotasPorStatus(notas.filter((item) => item.id !== nota.id)));
        this.confirmacaoExclusaoNota.set(null);
        this.mensagemSucesso.set(response?.mensagem ?? 'Nota fiscal deletada com sucesso.');
      },
      error: (err) => {
        this.mensagemErro.set(this.extrairMensagemErro(err, 'Nao foi possivel deletar a nota fiscal.'));
      },
      complete: () => {
        this.notaEmExclusaoId.set(null);
      },
    });
  }

  obterDescricaoProduto(codigo: number): string {
    return this.produtosDisponiveis().find((item) => item.codigo === codigo)?.descricao ?? `Produto #${codigo}`;
  }

  textoBotaoImpressao(nota: NotaFiscal): string {
    return this.notaEmImpressaoId() === nota.id ? 'Imprimindo...' : 'Imprimir';
  }

  textoBotaoCancelar(nota: NotaFiscal): string {
    return this.notaEmCancelamentoId() === nota.id ? 'Cancelando...' : 'Cancelar';
  }

  textoBotaoDeletar(nota: NotaFiscal): string {
    return this.notaEmExclusaoId() === nota.id ? 'Deletando...' : 'Deletar';
  }

  textoBotaoConfirmarExclusaoNota(): string {
    return this.notaEmExclusaoId() === null ? 'Excluir nota' : 'Excluindo...';
  }

  estaProcessandoNota(notaId: number): boolean {
    return (
      this.notaEmImpressaoId() === notaId ||
      this.notaEmCancelamentoId() === notaId ||
      this.notaEmExclusaoId() === notaId
    );
  }

  trackProduto(index: number, item: Produto): number | string {
    if (item.codigo > 0) {
      return item.codigo;
    }

    return `produto-${index}-${item.descricao}`;
  }

  trackItemNotaAtual(index: number, item: ItemNotaPayload): string {
    return `${item.produto_id}-${index}`;
  }

  trackNota(index: number, nota: NotaFiscal): number | string {
    if (nota.id > 0) {
      return nota.id;
    }

    return `nota-${index}-${nota.cliente}`;
  }

  trackItemNota(index: number, item: ItemNota): number | string {
    if (item.id > 0) {
      return item.id;
    }

    return `nota-item-${item.produto_id}-${index}`;
  }

  private limparMensagens(): void {
    this.mensagemErro.set('');
    this.mensagemSucesso.set('');
  }

  private normalizarProdutos(produtos: Produto[]): Produto[] {
    const produtosValidos = produtos.filter((item) => item.codigo > 0);
    const produtosUnicos = new Map<number, Produto>();

    for (const produto of produtosValidos) {
      if (!produtosUnicos.has(produto.codigo)) {
        produtosUnicos.set(produto.codigo, produto);
      }
    }

    return Array.from(produtosUnicos.values());
  }

  private ordenarNotasPorStatus(notas: NotaFiscal[]): NotaFiscal[] {
    return [...notas].sort((a, b) => {
      if (a.esta_aberta !== b.esta_aberta) {
        return a.esta_aberta ? -1 : 1;
      }

      return a.id - b.id;
    });
  }

  private extrairMensagemErro(err: unknown, fallback: string): string {
    const erroHttp = err as { error?: unknown };
    const resposta = erroHttp?.error;

    if (typeof resposta === 'string' && resposta.trim() !== '') {
      return resposta;
    }

    if (resposta && typeof resposta === 'object') {
      const respostaObj = resposta as Record<string, unknown>;

      for (const chave of ['erro', 'error', 'mensagem']) {
        const valor = respostaObj[chave];
        if (typeof valor === 'string' && valor.trim() !== '') {
          return valor;
        }
      }
    }

    return fallback;
  }
}
