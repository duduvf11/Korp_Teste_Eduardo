package db

import (
	"log"
	"os"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/joho/godotenv"

	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/models"
)

var DB *gorm.DB

func ConectaComBancoDeDados() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	dsn := os.Getenv("DB_URL")

	banco, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("Erro ao conectar com o banco de dados:", err)
	}

	banco.AutoMigrate(&models.Produto{})
	
	DB = banco
}