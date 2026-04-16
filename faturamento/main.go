package main

import (
	"github.com/gin-gonic/gin"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/db"
)

func main() {
	db.ConectarBancoDeDados()

	r := gin.Default()

	r.POST("/notas-fiscais", controllers.CriarNotaFiscal)
	r.GET("/notas-fiscais", controllers.ListarNotasFiscais)

	r.Run(":8081")
}