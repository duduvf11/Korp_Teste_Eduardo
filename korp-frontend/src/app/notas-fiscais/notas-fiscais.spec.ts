import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { NotasFiscais } from './notas-fiscais';

type ProdutoMock = {
  codigo: number;
  descricao: string;
  saldo: number;
  preco: number;
};

type ItemNotaPayloadMock = {
  produto_id: number;
  quantidade: number;
};

type ItemNotaMock = {
  id: number;
  produto_id: number;
  nota_fiscal_id: number;
  quantidade: number;
};

type NotaFiscalMock = {
  id: number;
  cliente: string;
  esta_aberta: boolean;
  itens: ItemNotaMock[];
};

function criarProduto(codigo: number, descricao: string, saldo: number, preco: number): ProdutoMock {
  return { codigo, descricao, saldo, preco };
}

function criarNota(id: number, aberta: boolean, cliente: string, itens: ItemNotaMock[] = []): NotaFiscalMock {
  return { id, cliente, esta_aberta: aberta, itens };
}

describe('NotasFiscais', () => {
  let component: NotasFiscais;
  let fixture: ComponentFixture<NotasFiscais>;
  let httpMock: HttpTestingController;

  const urlProdutos = 'http://localhost:8080/produtos';
  const urlNotas = 'http://localhost:8081/notas-fiscais';

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotasFiscais],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function iniciarComponente(produtosIniciais: ProdutoMock[], notasIniciais: NotaFiscalMock[]): void {
    fixture = TestBed.createComponent(NotasFiscais);
    component = fixture.componentInstance;
    fixture.detectChanges();

    const requisicaoProdutos = httpMock.expectOne(urlProdutos);
    expect(requisicaoProdutos.request.method).toBe('GET');
    requisicaoProdutos.flush(produtosIniciais);

    const requisicaoNotas = httpMock.expectOne(urlNotas);
    expect(requisicaoNotas.request.method).toBe('GET');
    requisicaoNotas.flush(notasIniciais);

    fixture.detectChanges();
  }

  it('should create', () => {
    iniciarComponente([], []);

    expect(component).toBeTruthy();
  });

  it('should ordenar notas com abertas no topo e por numero ascendente', () => {
    iniciarComponente([criarProduto(1, 'Produto A', 10, 15)], [
      criarNota(4, false, 'Cliente D'),
      criarNota(3, true, 'Cliente C'),
      criarNota(1, true, 'Cliente A'),
      criarNota(2, false, 'Cliente B'),
    ]);

    expect(component.notasFiscais().map((nota) => nota.id)).toEqual([1, 3, 2, 4]);
  });

  it('should validar saldo ao adicionar item na nota', () => {
    iniciarComponente([criarProduto(10, 'Teclado', 2, 120)], []);

    component.produtoSelecionadoCodigo = 10;
    component.quantidadeSelecionada = 3;
    component.adicionarItem();

    expect(component.itensNotaAtual()).toEqual([]);
    expect(component.mensagemErro()).toContain('Quantidade maior que o saldo disponivel');

    component.quantidadeSelecionada = 2;
    component.adicionarItem();

    const itensEsperados: ItemNotaPayloadMock[] = [{ produto_id: 10, quantidade: 2 }];
    expect(component.itensNotaAtual()).toEqual(itensEsperados);
    expect(component.mensagemSucesso()).toContain('Item adicionado na nota fiscal');
  });

  it('should salvar nota fiscal e recarregar lista de notas', () => {
    const produto = criarProduto(10, 'Teclado', 2, 120);
    const itensNota: ItemNotaPayloadMock[] = [{ produto_id: 10, quantidade: 1 }];
    iniciarComponente([produto], []);

    component.clienteNotaAtual = 'Empresa ABC';
    component.itensNotaAtual.set(itensNota);
    component.salvarNotaFiscal();

    const requisicaoSalvar = httpMock.expectOne(urlNotas);
    expect(requisicaoSalvar.request.method).toBe('POST');
    expect(requisicaoSalvar.request.body).toEqual({
      cliente: 'Empresa ABC',
      itens: itensNota,
    });
    requisicaoSalvar.flush(criarNota(7, true, 'Empresa ABC', [
      { id: 1, produto_id: 10, nota_fiscal_id: 7, quantidade: 1 },
    ]));

    const requisicaoRecarga = httpMock.expectOne(urlNotas);
    expect(requisicaoRecarga.request.method).toBe('GET');
    requisicaoRecarga.flush([
      criarNota(7, true, 'Empresa ABC', [
        { id: 1, produto_id: 10, nota_fiscal_id: 7, quantidade: 1 },
      ]),
    ]);

    expect(component.clienteNotaAtual).toBe('');
    expect(component.itensNotaAtual()).toEqual([]);
    expect(component.mensagemSucesso()).toContain('Nota fiscal criada com sucesso');
  });

  it('should cancelar nota aberta e atualizar o status para fechada', () => {
    const notaAberta = criarNota(5, true, 'Cliente XPTO');
    iniciarComponente([criarProduto(1, 'Produto', 10, 20)], [notaAberta]);

    component.cancelarNota(notaAberta);

    const requisicaoCancelar = httpMock.expectOne(`${urlNotas}/5/cancelar`);
    expect(requisicaoCancelar.request.method).toBe('POST');
    requisicaoCancelar.flush({
      mensagem: 'Nota fiscal cancelada com sucesso.',
      nota: criarNota(5, false, 'Cliente XPTO'),
    });

    expect(component.notasFiscais()[0].esta_aberta).toBe(false);
    expect(component.mensagemSucesso()).toContain('Nota fiscal cancelada com sucesso');
  });

  it('should excluir nota fechada apos confirmacao', () => {
    const notaFechada = criarNota(8, false, 'Cliente Final');
    iniciarComponente([criarProduto(1, 'Produto', 10, 20)], [notaFechada]);

    component.solicitarExclusaoNota(notaFechada);
    expect(component.confirmacaoExclusaoNota()?.id).toBe(8);

    component.confirmarExclusaoNota();

    const requisicaoExcluir = httpMock.expectOne(`${urlNotas}/8`);
    expect(requisicaoExcluir.request.method).toBe('DELETE');
    requisicaoExcluir.flush({
      mensagem: 'Nota fiscal deletada com sucesso.',
      nota_id: 8,
    });

    expect(component.notasFiscais()).toEqual([]);
    expect(component.confirmacaoExclusaoNota()).toBeNull();
    expect(component.notaEmExclusaoId()).toBeNull();
    expect(component.mensagemSucesso()).toContain('Nota fiscal deletada com sucesso');
  });
});
