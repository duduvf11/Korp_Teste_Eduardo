package main

import (
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConectaComBancoDeDados()

	r := gin.Default()

	r.POST("/produtos", controllers.CriarProduto)
	r.GET("/produtos", controllers.ListarProdutos)
	
	r.Run(":8080")
}