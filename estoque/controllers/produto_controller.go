package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RequisicaoBaixa struct {
	Quantidade int `json:"quantidade"`
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
	resposta := gin.H{
		"codigo": codigo,
		"erro":   mensagem,
	}

	if detalhes != nil {
		resposta["detalhes"] = detalhes
	}

	c.JSON(status, resposta)
}

func CriarProduto(c *gin.Context) {
	var produto models.Produto

	if err := c.ShouldBindJSON(&produto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	db.DB.Create(&produto)
	c.JSON(http.StatusCreated, produto)
}

func ListarProdutos(c *gin.Context) {
	var produtos []models.Produto

	db.DB.Find(&produtos)
	c.JSON(http.StatusOK, produtos)
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

		errBusca := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("codigo = ?", id).First(&produto).Error
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
			Where("codigo = ?", produto.Codigo).
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
