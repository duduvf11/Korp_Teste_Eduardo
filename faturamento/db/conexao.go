package db

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/models"
)

var DB *gorm.DB

func ConectarBancoDeDados() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	dsn := os.Getenv("DB_URL")

	banco, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("Erro ao conectar com o banco de dados:", err)
	}

	banco.AutoMigrate(&models.NotaFiscal{}, &models.ItemNota{})

	DB = banco
}