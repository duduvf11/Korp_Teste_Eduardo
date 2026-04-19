package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/estoque/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	rotaProdutosTeste            = "/produtos"
	formatoStatusInesperadoTeste = "status inesperado: esperado %d, obtido %d. body=%s"
	formatoErroInserirBaseTeste  = "erro ao inserir produto base: %v"
)

func setupRouterEstoque(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	banco, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("erro ao abrir banco em memoria: %v", err)
	}

	if err := banco.AutoMigrate(&models.Produto{}); err != nil {
		t.Fatalf("erro ao migrar banco em memoria: %v", err)
	}

	db.DB = banco

	router := gin.New()
	router.POST(rotaProdutosTeste, controllers.CriarProduto)
	router.GET(rotaProdutosTeste, controllers.ListarProdutos)
	router.PUT("/produtos/:id", controllers.AtualizarProduto)
	router.DELETE("/produtos/:id", controllers.DeletarProduto)
	router.POST("/produtos/:id/baixa", controllers.BaixarEstoque)

	return router
}

func performJSONRequest(t *testing.T, router *gin.Engine, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Buffer
	if payload == nil {
		body = bytes.NewBuffer(nil)
	} else {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("erro ao serializar payload: %v", err)
		}
		body = bytes.NewBuffer(data)
	}

	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func decodeJSON[T any](t *testing.T, res *httptest.ResponseRecorder) T {
	t.Helper()

	var out T
	if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
		t.Fatalf("erro ao desserializar resposta (%s): %v", res.Body.String(), err)
	}

	return out
}

func TestCriarProdutoRota(t *testing.T) {
	router := setupRouterEstoque(t)

	res := performJSONRequest(t, router, http.MethodPost, rotaProdutosTeste, map[string]any{
		"codigo":    10,
		"descricao": "Teclado",
		"saldo":     5,
		"preco":     129.9,
	})

	if res.Code != http.StatusCreated {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusCreated, res.Code, res.Body.String())
	}

	produto := decodeJSON[models.Produto](t, res)
	if produto.Codigo != 10 || produto.Descricao != "Teclado" || produto.Saldo != 5 {
		t.Fatalf("produto retornado invalido: %+v", produto)
	}
}

func TestListarProdutosRota(t *testing.T) {
	router := setupRouterEstoque(t)

	if err := db.DB.Create(&models.Produto{Codigo: 2, Descricao: "B", Saldo: 2, Preco: 20}).Error; err != nil {
		t.Fatalf("erro ao inserir produto 2: %v", err)
	}
	if err := db.DB.Create(&models.Produto{Codigo: 1, Descricao: "A", Saldo: 1, Preco: 10}).Error; err != nil {
		t.Fatalf("erro ao inserir produto 1: %v", err)
	}

	res := performJSONRequest(t, router, http.MethodGet, rotaProdutosTeste, nil)
	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	produtos := decodeJSON[[]models.Produto](t, res)
	if len(produtos) != 2 {
		t.Fatalf("quantidade inesperada de produtos: %d", len(produtos))
	}

	if produtos[0].Codigo != 1 || produtos[1].Codigo != 2 {
		t.Fatalf("ordenacao inesperada: %+v", produtos)
	}
}

func TestAtualizarProdutoRota(t *testing.T) {
	router := setupRouterEstoque(t)

	if err := db.DB.Create(&models.Produto{Codigo: 7, Descricao: "Mouse", Saldo: 4, Preco: 80}).Error; err != nil {
		t.Fatalf(formatoErroInserirBaseTeste, err)
	}

	res := performJSONRequest(t, router, http.MethodPut, "/produtos/7", map[string]any{
		"descricao": "Mouse sem fio",
		"saldo":     9,
		"preco":     95.5,
	})

	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	produto := decodeJSON[models.Produto](t, res)
	if produto.Codigo != 7 || produto.Descricao != "Mouse sem fio" || produto.Saldo != 9 {
		t.Fatalf("produto atualizado invalido: %+v", produto)
	}
}

func TestDeletarProdutoRota(t *testing.T) {
	router := setupRouterEstoque(t)

	if err := db.DB.Create(&models.Produto{Codigo: 3, Descricao: "Headset", Saldo: 3, Preco: 199}).Error; err != nil {
		t.Fatalf(formatoErroInserirBaseTeste, err)
	}

	res := performJSONRequest(t, router, http.MethodDelete, "/produtos/3", nil)
	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	var produto models.Produto
	err := db.DB.Where("codigo = ?", 3).First(&produto).Error
	if err == nil {
		t.Fatalf("produto nao foi removido do banco")
	}
}

func TestBaixarEstoqueRota(t *testing.T) {
	router := setupRouterEstoque(t)

	if err := db.DB.Create(&models.Produto{Codigo: 11, Descricao: "Webcam", Saldo: 10, Preco: 300}).Error; err != nil {
		t.Fatalf(formatoErroInserirBaseTeste, err)
	}

	res := performJSONRequest(t, router, http.MethodPost, "/produtos/11/baixa", map[string]any{
		"quantidade": 3,
	})

	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	var produto models.Produto
	if err := db.DB.Where("codigo = ?", 11).First(&produto).Error; err != nil {
		t.Fatalf("erro ao buscar produto atualizado: %v", err)
	}

	if produto.Saldo != 7 {
		t.Fatalf("saldo inesperado apos baixa: esperado 7, obtido %d", produto.Saldo)
	}
}
