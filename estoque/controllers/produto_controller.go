package controllers

import (
	"net/http"

	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/models"
	"github.com/gin-gonic/gin"
)

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