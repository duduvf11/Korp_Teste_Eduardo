package main

import (
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConectarBancoDeDados()
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:4200"}
	r.Use(cors.New(config))

	r.POST("/notas-fiscais", controllers.CriarNotaFiscal)
	r.GET("/notas-fiscais", controllers.ListarNotasFiscais)
	r.POST("/notas-fiscais/:id/imprimir", controllers.ImprimirNota)
	r.POST("/notas-fiscais/:id/cancelar", controllers.CancelarNota)
	r.DELETE("/notas-fiscais/:id", controllers.DeletarNota)

	r.Run(":8081")
}
