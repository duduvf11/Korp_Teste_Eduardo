package controllers

import (
	"errors"
    "net/http"
    "strconv"

    "github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
    "github.com/duduvf11/Korp_Teste_Eduardo/estoque/models"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type RequisicaoBaixa struct {
	Quantidade int `json:"quantidade"`
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
        c.JSON(http.StatusBadRequest, gin.H{"erro": "Formato de dados inválido"})
        return
    }

    if req.Quantidade <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"erro": "Quantidade deve ser maior que zero"})
        return
    }

    idTexto := c.Param("id")
    id, err := strconv.Atoi(idTexto)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"erro": "ID do produto inválido"})
        return
    }

    var produto models.Produto
    resultado := db.DB.First(&produto, id)

    if errors.Is(resultado.Error, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"erro": "Produto não encontrado"})
        return
    }

    if resultado.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao consultar produto no banco"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "mensagem":             "Produto localizado. Próxima etapa: validar saldo e atualizar estoque.",
        "produto_codigo":       produto.Codigo,
        "saldo_atual":          produto.Saldo,
        "quantidade_solicitada": req.Quantidade,
    })
}