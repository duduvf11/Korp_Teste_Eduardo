package main

import (
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConectaComBancoDeDados()
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:4200"}
	r.Use(cors.New(config))

	r.POST("/produtos", controllers.CriarProduto)
	r.GET("/produtos", controllers.ListarProdutos)
	r.PUT("/produtos/:id", controllers.AtualizarProduto)
	r.DELETE("/produtos/:id", controllers.DeletarProduto)

	r.POST("/produtos/:id/baixa", controllers.BaixarEstoque)

	r.Run(":8080")
}
