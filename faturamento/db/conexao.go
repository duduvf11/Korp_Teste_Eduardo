package db

import (
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/models"
)

var DB *gorm.DB

func ConectarBancoDeDados() {
	dsn := "host=localhost user=postgres password=Andrea021025@ dbname=faturamento_db port=5433 sslmode=disable"

	banco, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("Erro ao conectar com o banco de dados:", err)
	}

	banco.AutoMigrate(&models.NotaFiscal{}, &models.ItemNota{})

	DB = banco
}