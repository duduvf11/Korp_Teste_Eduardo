import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { Produtos } from './produtos';
import { environment } from '../../environments/environment';

type ProdutoMock = {
  codigo: number;
  descricao: string;
  saldo: number;
  preco: number;
};

function criarProduto(codigo: number, descricao: string, saldo: number, preco: number): ProdutoMock {
  return { codigo, descricao, saldo, preco };
}

describe('Produtos', () => {
  let component: Produtos;
  let fixture: ComponentFixture<Produtos>;
  let httpMock: HttpTestingController;

  const urlProdutos = `${environment.api.estoqueBaseUrl}/produtos`;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Produtos],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function iniciarComponente(produtosIniciais: ProdutoMock[]): void {
    fixture = TestBed.createComponent(Produtos);
    component = fixture.componentInstance;
    fixture.detectChanges();

    const requisicaoInicial = httpMock.expectOne(urlProdutos);
    expect(requisicaoInicial.request.method).toBe('GET');
    requisicaoInicial.flush(produtosIniciais);

    fixture.detectChanges();
  }

  it('should create', () => {
    iniciarComponente([]);

    expect(component).toBeTruthy();
  });

  it('should carregar produtos ordenados por codigo', () => {
    iniciarComponente([
      criarProduto(3, 'Produto C', 3, 30),
      criarProduto(1, 'Produto A', 1, 10),
      criarProduto(2, 'Produto B', 2, 20),
    ]);

    expect(component.listaProdutos().map((produto) => produto.codigo)).toEqual([1, 2, 3]);
  });

  it('should salvar novo produto e recarregar a lista', () => {
    iniciarComponente([]);

    component.produto = criarProduto(10, '  Teclado Gamer  ', 7, 199.9);
    component.salvarProduto();

    const requisicaoSalvar = httpMock.expectOne(urlProdutos);
    expect(requisicaoSalvar.request.method).toBe('POST');
    expect(requisicaoSalvar.request.body).toEqual(criarProduto(10, 'Teclado Gamer', 7, 199.9));
    requisicaoSalvar.flush(criarProduto(10, 'Teclado Gamer', 7, 199.9));

    const requisicaoRecarga = httpMock.expectOne(urlProdutos);
    expect(requisicaoRecarga.request.method).toBe('GET');
    requisicaoRecarga.flush([criarProduto(10, 'Teclado Gamer', 7, 199.9)]);

    expect(component.mensagemSucesso()).toContain('Produto salvo com sucesso');
    expect(component.modoEdicao()).toBe(false);
    expect(component.produto.codigo).toBe(0);
  });

  it('should atualizar produto em modo de edicao', () => {
    const produtoAtual = criarProduto(22, 'Mouse', 5, 89.9);
    iniciarComponente([produtoAtual]);

    component.iniciarEdicao(produtoAtual);
    component.produto.descricao = '  Mouse sem fio  ';
    component.salvarProduto();

    const requisicaoAtualizar = httpMock.expectOne(`${urlProdutos}/22`);
    expect(requisicaoAtualizar.request.method).toBe('PUT');
    expect(requisicaoAtualizar.request.body).toEqual(criarProduto(22, 'Mouse sem fio', 5, 89.9));
    requisicaoAtualizar.flush(criarProduto(22, 'Mouse sem fio', 5, 89.9));

    const requisicaoRecarga = httpMock.expectOne(urlProdutos);
    expect(requisicaoRecarga.request.method).toBe('GET');
    requisicaoRecarga.flush([criarProduto(22, 'Mouse sem fio', 5, 89.9)]);

    expect(component.mensagemSucesso()).toContain('Produto atualizado com sucesso');
    expect(component.modoEdicao()).toBe(false);
  });

  it('should excluir produto confirmado e atualizar a lista', () => {
    const produto = criarProduto(9, 'Hub USB', 2, 49.9);
    iniciarComponente([produto]);

    component.solicitarExclusaoProduto(produto);
    expect(component.confirmacaoExclusaoProduto()?.codigo).toBe(9);

    component.confirmarExclusaoProduto();

    const requisicaoExcluir = httpMock.expectOne(`${urlProdutos}/9`);
    expect(requisicaoExcluir.request.method).toBe('DELETE');
    requisicaoExcluir.flush({ mensagem: 'Produto deletado com sucesso.' });

    const requisicaoRecarga = httpMock.expectOne(urlProdutos);
    expect(requisicaoRecarga.request.method).toBe('GET');
    requisicaoRecarga.flush([]);

    expect(component.confirmacaoExclusaoProduto()).toBeNull();
    expect(component.produtoEmExclusaoCodigo()).toBeNull();
    expect(component.listaProdutos()).toEqual([]);
    expect(component.mensagemSucesso()).toContain('Produto deletado com sucesso');
  });
});
