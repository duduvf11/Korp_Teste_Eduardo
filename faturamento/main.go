package main

import (
	"log"
	"os"
	"strings"

	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConectarBancoDeDados()
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.RequestContextLogger("faturamento"))

	config := cors.DefaultConfig()
	config.AllowOrigins = obterOrigensPermitidas()
	r.Use(cors.New(config))

	log.Printf("service=faturamento level=info message=%q allowed_origins=%v", "iniciando servico", config.AllowOrigins)

	r.POST("/notas-fiscais", controllers.CriarNotaFiscal)
	r.GET("/notas-fiscais", controllers.ListarNotasFiscais)
	r.POST("/notas-fiscais/:id/imprimir", controllers.ImprimirNota)
	r.POST("/notas-fiscais/:id/cancelar", controllers.CancelarNota)
	r.DELETE("/notas-fiscais/:id", controllers.DeletarNota)

	r.Run(":8081")
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
