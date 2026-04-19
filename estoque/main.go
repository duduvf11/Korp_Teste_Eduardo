package main

import (
	"log"
	"os"
	"strings"

	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConectaComBancoDeDados()
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.RequestContextLogger("estoque"))

	config := cors.DefaultConfig()
	config.AllowOrigins = obterOrigensPermitidas()
	r.Use(cors.New(config))

	log.Printf("service=estoque level=info message=%q allowed_origins=%v", "iniciando servico", config.AllowOrigins)

	r.POST("/produtos", controllers.CriarProduto)
	r.GET("/produtos", controllers.ListarProdutos)
	r.PUT("/produtos/:id", controllers.AtualizarProduto)
	r.DELETE("/produtos/:id", controllers.DeletarProduto)

	r.POST("/produtos/:id/baixa", controllers.BaixarEstoque)

	r.Run(":8080")
}

func obterOrigensPermitidas() []string {
	origensBrutas := strings.TrimSpace(os.Getenv("ALLOW_ORIGINS"))
	if origensBrutas == "" {
		return []string{"http://localhost:4200"}
	}

	partes := strings.Split(origensBrutas, ",")
	origens := make([]string, 0, len(partes))
	for _, parte := range partes {
		origem := strings.TrimSpace(parte)
		if origem != "" {
			origens = append(origens, origem)
		}
	}

	if len(origens) == 0 {
		return []string{"http://localhost:4200"}
	}

	return origens
}
