import { expect, test, type Page, type Route } from '@playwright/test';

type Produto = {
  codigo: number;
  descricao: string;
  saldo: number;
  preco: number;
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

type EstadoProdutos = {
  produtos: Produto[];
};

type EstadoNotas = {
  notas: NotaFiscal[];
  proximoID: number;
  falharImpressao: boolean;
  mensagemFalhaImpressao: string;
};

const API_ESTOQUE = 'http://localhost:8080';
const API_FATURAMENTO = 'http://localhost:8081';

function responderJSON(route: Route, status: number, payload: unknown): Promise<void> {
  return route.fulfill({
    status,
    contentType: 'application/json',
    body: JSON.stringify(payload),
  });
}

function extrairID(urlCompleta: string, segmento: 'produtos' | 'notas-fiscais'): number {
  const pathname = new URL(urlCompleta).pathname;
  const partes = pathname.split('/').filter(Boolean);
  const indice = partes.indexOf(segmento);

  if (indice < 0 || indice + 1 >= partes.length) {
    return 0;
  }

  return Number(partes[indice + 1]);
}

async function mockApiProdutos(page: Page, estado: EstadoProdutos): Promise<void> {
  await page.route(`${API_ESTOQUE}/produtos`, async (route) => {
    const requisicao = route.request();
    const method = requisicao.method();

    if (method === 'GET') {
      const produtosOrdenados = [...estado.produtos].sort((a, b) => a.codigo - b.codigo);
      await responderJSON(route, 200, produtosOrdenados);
      return;
    }

    if (method === 'POST') {
      const body = requisicao.postDataJSON() as Produto;
      estado.produtos.push({ ...body });
      await responderJSON(route, 201, body);
      return;
    }

    await route.abort();
  });

  await page.route(`${API_ESTOQUE}/produtos/*`, async (route) => {
    const requisicao = route.request();
    const method = requisicao.method();
    const idProduto = extrairID(requisicao.url(), 'produtos');

    if (!idProduto) {
      await responderJSON(route, 400, { erro: 'Produto invalido' });
      return;
    }

    if (method === 'PUT') {
      const body = requisicao.postDataJSON() as Produto;
      estado.produtos = estado.produtos.map((produto) => (produto.codigo === idProduto ? { ...body } : produto));
      await responderJSON(route, 200, body);
      return;
    }

    if (method === 'DELETE') {
      estado.produtos = estado.produtos.filter((produto) => produto.codigo !== idProduto);
      await responderJSON(route, 200, {
        mensagem: 'Produto deletado com sucesso.',
        codigo: idProduto,
      });
      return;
    }

    await route.abort();
  });
}

async function mockApiNotas(page: Page, estadoNotas: EstadoNotas, estadoProdutos: EstadoProdutos): Promise<void> {
  await page.route(`${API_FATURAMENTO}/notas-fiscais`, async (route) => {
    const requisicao = route.request();

    if (requisicao.method() === 'GET') {
      await responderJSON(route, 200, estadoNotas.notas);
      return;
    }

    if (requisicao.method() === 'POST') {
      const body = requisicao.postDataJSON() as { cliente: string; itens: Array<{ produto_id: number; quantidade: number }> };
      const notaID = estadoNotas.proximoID;
      estadoNotas.proximoID += 1;

      const itens: ItemNota[] = body.itens.map((item, index) => ({
        id: index + 1,
        produto_id: item.produto_id,
        nota_fiscal_id: notaID,
        quantidade: item.quantidade,
      }));

      const notaCriada: NotaFiscal = {
        id: notaID,
        cliente: body.cliente,
        esta_aberta: true,
        itens,
      };

      estadoNotas.notas.push(notaCriada);
      await responderJSON(route, 201, notaCriada);
      return;
    }

    await route.abort();
  });

  await page.route(`${API_FATURAMENTO}/notas-fiscais/*/imprimir`, async (route) => {
    const notaID = extrairID(route.request().url(), 'notas-fiscais');
    const nota = estadoNotas.notas.find((item) => item.id === notaID);

    if (!nota) {
      await responderJSON(route, 404, { codigo: 'NOTA_NAO_ENCONTRADA', erro: 'Nota nao encontrada.' });
      return;
    }

    if (estadoNotas.falharImpressao) {
      await responderJSON(route, 503, {
        codigo: 'ESTOQUE_INDISPONIVEL',
        erro: estadoNotas.mensagemFalhaImpressao,
      });
      return;
    }

    nota.esta_aberta = false;

    for (const item of nota.itens) {
      estadoProdutos.produtos = estadoProdutos.produtos.map((produto) => {
        if (produto.codigo !== item.produto_id) {
          return produto;
        }

        return {
          ...produto,
          saldo: Math.max(0, produto.saldo - item.quantidade),
        };
      });
    }

    await responderJSON(route, 200, {
      mensagem: 'Nota fiscal impressa e estoque atualizado com sucesso.',
      nota,
    });
  });

  await page.route(`${API_FATURAMENTO}/notas-fiscais/*/cancelar`, async (route) => {
    const notaID = extrairID(route.request().url(), 'notas-fiscais');
    const nota = estadoNotas.notas.find((item) => item.id === notaID);

    if (!nota) {
      await responderJSON(route, 404, { codigo: 'NOTA_NAO_ENCONTRADA', erro: 'Nota nao encontrada.' });
      return;
    }

    if (!nota.esta_aberta) {
      await responderJSON(route, 409, { codigo: 'NOTA_FECHADA', erro: 'Apenas notas em aberto podem ser canceladas.' });
      return;
    }

    nota.esta_aberta = false;

    await responderJSON(route, 200, {
      mensagem: 'Nota fiscal cancelada com sucesso.',
      nota,
    });
  });

  await page.route(`${API_FATURAMENTO}/notas-fiscais/*`, async (route) => {
    const requisicao = route.request();
    if (requisicao.method() !== 'DELETE') {
      await route.abort();
      return;
    }

    const notaID = extrairID(requisicao.url(), 'notas-fiscais');
    const nota = estadoNotas.notas.find((item) => item.id === notaID);

    if (!nota) {
      await responderJSON(route, 404, { codigo: 'NOTA_NAO_ENCONTRADA', erro: 'Nota nao encontrada.' });
      return;
    }

    if (nota.esta_aberta) {
      await responderJSON(route, 409, { codigo: 'NOTA_ABERTA', erro: 'A nota ainda esta aberta.' });
      return;
    }

    estadoNotas.notas = estadoNotas.notas.filter((item) => item.id !== notaID);

    await responderJSON(route, 200, {
      mensagem: 'Nota fiscal deletada com sucesso.',
      nota_id: notaID,
    });
  });
}

async function selecionarProdutoPorDescricao(page: Page, descricao: string): Promise<void> {
  const selectProduto = page.getByTestId('nota-produto-select');
  await expect(selectProduto).toContainText(descricao);

  await selectProduto.evaluate((elemento, descricaoProduto) => {
    const select = elemento as HTMLSelectElement;
    const opcao = Array.from(select.options).find((item) =>
      item.textContent?.toLowerCase().includes(String(descricaoProduto).toLowerCase()),
    );

    if (!opcao) {
      throw new Error(`Opcao de produto ${descricaoProduto} nao encontrada.`);
    }

    select.value = opcao.value;
    select.dispatchEvent(new Event('change', { bubbles: true }));
  }, descricao);
}

test('produtos: criar, editar e excluir com confirmacao', async ({ page }) => {
  const estadoProdutos: EstadoProdutos = { produtos: [] };
  await mockApiProdutos(page, estadoProdutos);

  await page.goto('/produtos');

  await page.getByTestId('produto-codigo-input').fill('101');
  await page.getByTestId('produto-descricao-input').fill('Teclado Gamer');
  await page.getByTestId('produto-saldo-input').fill('7');
  await page.getByTestId('produto-preco-input').fill('199.90');
  await page.getByTestId('produto-salvar-btn').click();

  await expect(page.getByTestId('produtos-mensagem-sucesso')).toContainText('Produto salvo com sucesso');
  await expect(page.getByTestId('produto-row-101')).toContainText('Teclado Gamer');

  await page.getByTestId('produto-editar-101').click();
  await page.getByTestId('produto-descricao-input').fill('Teclado Gamer RGB');
  await page.getByTestId('produto-salvar-btn').click();

  await expect(page.getByTestId('produtos-mensagem-sucesso')).toContainText('Produto atualizado com sucesso');
  await expect(page.getByTestId('produto-row-101')).toContainText('Teclado Gamer RGB');

  await page.getByTestId('produto-deletar-101').click();
  await expect(page.getByTestId('produto-confirmacao-modal')).toBeVisible();
  await page.getByTestId('produto-confirmacao-confirmar').click();

  await expect(page.getByTestId('produtos-mensagem-sucesso')).toContainText('Produto deletado com sucesso');
  await expect(page.getByTestId('produto-row-101')).toHaveCount(0);
});

test('notas: criar, imprimir, cancelar e excluir por status', async ({ page }) => {
  const estadoProdutos: EstadoProdutos = {
    produtos: [
      { codigo: 10, descricao: 'Teclado', saldo: 9, preco: 120 },
      { codigo: 11, descricao: 'Mouse', saldo: 5, preco: 80 },
    ],
  };

  const estadoNotas: EstadoNotas = {
    notas: [],
    proximoID: 1,
    falharImpressao: false,
    mensagemFalhaImpressao: 'Falha inesperada na impressao.',
  };

  await mockApiProdutos(page, estadoProdutos);
  await mockApiNotas(page, estadoNotas, estadoProdutos);

  await page.goto('/notas-fiscais');

  await page.getByTestId('nota-cliente-input').fill('Cliente XPTO');
  await selecionarProdutoPorDescricao(page, 'Teclado');
  await page.getByTestId('nota-quantidade-input').fill('2');
  await page.getByTestId('nota-adicionar-item-btn').click();
  await page.getByTestId('nota-salvar-btn').click();

  await expect(page.getByTestId('notas-mensagem-sucesso')).toContainText('Nota fiscal criada com sucesso');
  await expect(page.getByTestId('nota-status-1')).toHaveText('Aberta');

  await page.getByTestId('nota-imprimir-1').click();
  await expect(page.getByTestId('notas-mensagem-sucesso')).toContainText('Nota fiscal impressa');
  await expect(page.getByTestId('nota-status-1')).toHaveText('Fechada');

  await page.getByTestId('nota-cliente-input').fill('Cliente Secundario');
  await selecionarProdutoPorDescricao(page, 'Mouse');
  await page.getByTestId('nota-quantidade-input').fill('1');
  await page.getByTestId('nota-adicionar-item-btn').click();
  await page.getByTestId('nota-salvar-btn').click();

  await expect(page.getByTestId('nota-status-2')).toHaveText('Aberta');

  await page.getByTestId('nota-cancelar-2').click();
  await expect(page.getByTestId('notas-mensagem-sucesso')).toContainText('Nota fiscal cancelada com sucesso');
  await expect(page.getByTestId('nota-status-2')).toHaveText('Fechada');

  await page.getByTestId('nota-deletar-1').click();
  await expect(page.getByTestId('nota-confirmacao-modal')).toBeVisible();
  await page.getByTestId('nota-confirmacao-confirmar').click();

  await expect(page.getByTestId('notas-mensagem-sucesso')).toContainText('Nota fiscal deletada com sucesso');
  await expect(page.getByTestId('nota-row-1')).toHaveCount(0);
});

test('notas: falha de impressao mostra erro amigavel e mantem nota aberta', async ({ page }) => {
  const estadoProdutos: EstadoProdutos = {
    produtos: [{ codigo: 20, descricao: 'Monitor', saldo: 3, preco: 799 }],
  };

  const estadoNotas: EstadoNotas = {
    notas: [
      {
        id: 55,
        cliente: 'Cliente Falha',
        esta_aberta: true,
        itens: [{ id: 1, produto_id: 20, nota_fiscal_id: 55, quantidade: 1 }],
      },
    ],
    proximoID: 56,
    falharImpressao: true,
    mensagemFalhaImpressao: 'O servico de estoque esta temporariamente indisponivel.',
  };

  await mockApiProdutos(page, estadoProdutos);
  await mockApiNotas(page, estadoNotas, estadoProdutos);

  await page.goto('/notas-fiscais');

  await expect(page.getByTestId('nota-status-55')).toHaveText('Aberta');
  await page.getByTestId('nota-imprimir-55').click();

  await expect(page.getByTestId('notas-mensagem-erro')).toContainText('temporariamente indisponivel');
  await expect(page.getByTestId('nota-status-55')).toHaveText('Aberta');
});
