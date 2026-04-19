package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const filtroCodigo = "codigo = ?"

type RequisicaoBaixa struct {
	Quantidade int `json:"quantidade"`
}

type RequisicaoAtualizacaoProduto struct {
	Codigo    *int    `json:"codigo"`
	Descricao string  `json:"descricao"`
	Saldo     int     `json:"saldo"`
	Preco     float64 `json:"preco"`
}

type ErroBaixaEstoque struct {
	StatusCode int
	Codigo     string
	Mensagem   string
	Detalhes   any
}

func (e *ErroBaixaEstoque) Error() string {
	return e.Mensagem
}

func responderErro(c *gin.Context, status int, codigo, mensagem string, detalhes any) {
	c.Set("erro_codigo", codigo)
	c.Set("erro_mensagem", mensagem)

	resposta := gin.H{
		"codigo": codigo,
		"erro":   mensagem,
	}

	if requestID, existe := c.Get("request_id"); existe {
		if requestIDTexto, ok := requestID.(string); ok && requestIDTexto != "" {
			resposta["request_id"] = requestIDTexto
		}
	}

	if detalhes != nil {
		resposta["detalhes"] = detalhes
	}

	c.JSON(status, resposta)
}

func CriarProduto(c *gin.Context) {
	var produto models.Produto

	if err := c.ShouldBindJSON(&produto); err != nil {
		responderErro(c, http.StatusBadRequest, "FORMATO_INVALIDO", "Dados invalidos para cadastro de produto.", err.Error())
		return
	}

	produto.Descricao = strings.TrimSpace(produto.Descricao)

	if produto.Codigo <= 0 {
		responderErro(c, http.StatusBadRequest, "CODIGO_INVALIDO", "O codigo do produto deve ser maior que zero.", nil)
		return
	}

	if produto.Descricao == "" {
		responderErro(c, http.StatusBadRequest, "DESCRICAO_OBRIGATORIA", "A descricao do produto e obrigatoria.", nil)
		return
	}

	if produto.Saldo < 0 {
		responderErro(c, http.StatusBadRequest, "SALDO_INVALIDO", "O saldo do produto nao pode ser negativo.", nil)
		return
	}

	if produto.Preco < 0 {
		responderErro(c, http.StatusBadRequest, "PRECO_INVALIDO", "O preco do produto nao pode ser negativo.", nil)
		return
	}

	var produtoExistente models.Produto
	errBusca := db.DB.Where(filtroCodigo, produto.Codigo).First(&produtoExistente).Error
	if errBusca == nil {
		responderErro(c, http.StatusConflict, "CODIGO_DUPLICADO", "Ja existe um produto cadastrado com esse codigo.", gin.H{"codigo": produto.Codigo})
		return
	}

	if !errors.Is(errBusca, gorm.ErrRecordNotFound) {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel validar o codigo do produto.", errBusca.Error())
		return
	}

	if err := db.DB.Create(&produto).Error; err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel salvar o produto.", err.Error())
		return
	}

	c.JSON(http.StatusCreated, produto)
}

func ListarProdutos(c *gin.Context) {
	var produtos []models.Produto

	db.DB.Order("codigo asc").Find(&produtos)
	c.JSON(http.StatusOK, produtos)
}

func AtualizarProduto(c *gin.Context) {
	idTexto := c.Param("id")
	codigoProduto, err := strconv.Atoi(idTexto)
	if err != nil || codigoProduto <= 0 {
		responderErro(c, http.StatusBadRequest, "CODIGO_INVALIDO", "Codigo do produto invalido.", nil)
		return
	}

	var req RequisicaoAtualizacaoProduto
	if err := c.ShouldBindJSON(&req); err != nil {
		responderErro(c, http.StatusBadRequest, "FORMATO_INVALIDO", "Dados invalidos para atualizacao de produto.", err.Error())
		return
	}

	if req.Codigo != nil && *req.Codigo != codigoProduto {
		responderErro(c, http.StatusBadRequest, "CODIGO_IMUTAVEL", "O codigo do produto nao pode ser alterado.", gin.H{"codigo_rota": codigoProduto, "codigo_payload": *req.Codigo})
		return
	}

	req.Descricao = strings.TrimSpace(req.Descricao)
	if req.Descricao == "" {
		responderErro(c, http.StatusBadRequest, "DESCRICAO_OBRIGATORIA", "A descricao do produto e obrigatoria.", nil)
		return
	}

	if req.Saldo < 0 {
		responderErro(c, http.StatusBadRequest, "SALDO_INVALIDO", "O saldo do produto nao pode ser negativo.", nil)
		return
	}

	if req.Preco < 0 {
		responderErro(c, http.StatusBadRequest, "PRECO_INVALIDO", "O preco do produto nao pode ser negativo.", nil)
		return
	}

	var produto models.Produto
	err = db.DB.Where(filtroCodigo, codigoProduto).First(&produto).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responderErro(c, http.StatusNotFound, "PRODUTO_NAO_ENCONTRADO", "Produto nao encontrado para atualizacao.", gin.H{"codigo": codigoProduto})
		return
	}

	if err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel consultar o produto para atualizacao.", err.Error())
		return
	}

	produto.Descricao = req.Descricao
	produto.Saldo = req.Saldo
	produto.Preco = req.Preco

	if err := db.DB.Model(&models.Produto{}).
		Where(filtroCodigo, codigoProduto).
		Updates(map[string]any{
			"descricao": produto.Descricao,
			"saldo":     produto.Saldo,
			"preco":     produto.Preco,
		}).Error; err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel atualizar o produto.", err.Error())
		return
	}

	c.JSON(http.StatusOK, produto)
}

func DeletarProduto(c *gin.Context) {
	idTexto := c.Param("id")
	codigoProduto, err := strconv.Atoi(idTexto)
	if err != nil || codigoProduto <= 0 {
		responderErro(c, http.StatusBadRequest, "CODIGO_INVALIDO", "Codigo do produto invalido.", nil)
		return
	}

	resultado := db.DB.Where(filtroCodigo, codigoProduto).Delete(&models.Produto{})
	if resultado.Error != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel deletar o produto.", resultado.Error.Error())
		return
	}

	if resultado.RowsAffected == 0 {
		responderErro(c, http.StatusNotFound, "PRODUTO_NAO_ENCONTRADO", "Produto nao encontrado para delecao.", gin.H{"codigo": codigoProduto})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem": "Produto deletado com sucesso.",
		"codigo":   codigoProduto,
	})
}

func BaixarEstoque(c *gin.Context) {
	var req RequisicaoBaixa
	if err := c.ShouldBindJSON(&req); err != nil {
		responderErro(c, http.StatusBadRequest, "FORMATO_INVALIDO", "Formato de dados invalido.", err.Error())
		return
	}

	if req.Quantidade <= 0 {
		responderErro(c, http.StatusBadRequest, "QUANTIDADE_INVALIDA", "Quantidade deve ser maior que zero.", nil)
		return
	}

	idTexto := c.Param("id")
	id, err := strconv.Atoi(idTexto)
	if err != nil || id <= 0 {
		responderErro(c, http.StatusBadRequest, "ID_INVALIDO", "ID do produto invalido.", nil)
		return
	}

	var saldoAnterior int
	var saldoNovo int

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var produto models.Produto

		errBusca := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(filtroCodigo, id).First(&produto).Error
		if errors.Is(errBusca, gorm.ErrRecordNotFound) {
			return &ErroBaixaEstoque{
				StatusCode: http.StatusNotFound,
				Codigo:     "PRODUTO_NAO_ENCONTRADO",
				Mensagem:   "Produto nao encontrado.",
				Detalhes: gin.H{
					"produto_codigo": id,
				},
			}
		}

		if errBusca != nil {
			return &ErroBaixaEstoque{
				StatusCode: http.StatusInternalServerError,
				Codigo:     "ERRO_BANCO",
				Mensagem:   "Erro ao consultar produto no banco.",
				Detalhes:   errBusca.Error(),
			}
		}

		if produto.Saldo < req.Quantidade {
			return &ErroBaixaEstoque{
				StatusCode: http.StatusConflict,
				Codigo:     "SALDO_INSUFICIENTE",
				Mensagem:   "Saldo insuficiente para realizar a baixa.",
				Detalhes: gin.H{
					"produto_codigo":        produto.Codigo,
					"saldo_atual":           produto.Saldo,
					"quantidade_solicitada": req.Quantidade,
				},
			}
		}

		saldoAnterior = produto.Saldo
		saldoNovo = produto.Saldo - req.Quantidade

		errAtualizacao := tx.Model(&models.Produto{}).
			Where(filtroCodigo, produto.Codigo).
			Update("saldo", saldoNovo).Error
		if errAtualizacao != nil {
			return &ErroBaixaEstoque{
				StatusCode: http.StatusInternalServerError,
				Codigo:     "ERRO_ATUALIZACAO_SALDO",
				Mensagem:   "Nao foi possivel atualizar o saldo do produto.",
				Detalhes:   errAtualizacao.Error(),
			}
		}

		return nil
	})

	if err != nil {
		var erroRegra *ErroBaixaEstoque
		if errors.As(err, &erroRegra) {
			responderErro(c, erroRegra.StatusCode, erroRegra.Codigo, erroRegra.Mensagem, erroRegra.Detalhes)
			return
		}

		responderErro(c, http.StatusInternalServerError, "ERRO_TRANSACAO", "Falha ao processar baixa de estoque.", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem":           "Baixa de estoque realizada com sucesso.",
		"produto_codigo":     id,
		"quantidade_baixada": req.Quantidade,
		"saldo_anterior":     saldoAnterior,
		"saldo_atual":        saldoNovo,
		"detalhe_calculo":    fmt.Sprintf("%d - %d = %d", saldoAnterior, req.Quantidade, saldoNovo),
	})
}
