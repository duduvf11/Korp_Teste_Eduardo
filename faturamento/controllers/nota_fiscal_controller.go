package controllers

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/models"
)

func CriarNotaFiscal(c *gin.Context) {
	var nota models.NotaFiscal
	if err := c.ShouldBindJSON(&nota); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.DB.Create(&nota)
	c.JSON(http.StatusCreated, nota)

}

func ListarNotasFiscais(c *gin.Context) {
	var notas []models.NotaFiscal

	db.DB.Find(&notas)
	c.JSON(http.StatusOK, notas)
}