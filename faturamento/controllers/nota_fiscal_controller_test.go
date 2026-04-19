package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/controllers"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	rotaNotasFiscaisTeste        = "/notas-fiscais"
	clientePadraoTeste           = "Cliente A"
	headerContentTypeTeste       = "Content-Type"
	valorJSONTeste               = "application/json"
	formatoStatusInesperadoTeste = "status inesperado: esperado %d, obtido %d. body=%s"
	formatoCodigoErroInesperado  = "codigo de erro inesperado: %+v"
)

func setupRouterFaturamento(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	banco, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("erro ao abrir banco em memoria: %v", err)
	}

	if err := banco.AutoMigrate(&models.NotaFiscal{}, &models.ItemNota{}); err != nil {
		t.Fatalf("erro ao migrar banco em memoria: %v", err)
	}

	db.DB = banco

	router := gin.New()
	router.POST(rotaNotasFiscaisTeste, controllers.CriarNotaFiscal)
	router.GET(rotaNotasFiscaisTeste, controllers.ListarNotasFiscais)
	router.POST("/notas-fiscais/:id/imprimir", controllers.ImprimirNota)
	router.POST("/notas-fiscais/:id/cancelar", controllers.CancelarNota)
	router.DELETE("/notas-fiscais/:id", controllers.DeletarNota)

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
		req.Header.Set(headerContentTypeTeste, valorJSONTeste)
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

func inserirNota(t *testing.T, cliente string, aberta bool, itens []models.ItemNota) models.NotaFiscal {
	t.Helper()

	nota := models.NotaFiscal{
		Cliente:    cliente,
		EstaAberta: aberta,
	}

	if err := db.DB.Create(&nota).Error; err != nil {
		t.Fatalf("erro ao criar nota no banco: %v", err)
	}

	for _, item := range itens {
		item.ID = 0
		item.NotaFiscalID = nota.ID
		if err := db.DB.Create(&item).Error; err != nil {
			t.Fatalf("erro ao criar item da nota no banco: %v", err)
		}
	}

	if err := db.DB.Preload("Item").First(&nota, nota.ID).Error; err != nil {
		t.Fatalf("erro ao recarregar nota no banco: %v", err)
	}

	return nota
}

func TestCriarNotaFiscalRota(t *testing.T) {
	router := setupRouterFaturamento(t)

	res := performJSONRequest(t, router, http.MethodPost, rotaNotasFiscaisTeste, map[string]any{
		"cliente": clientePadraoTeste,
		"itens": []map[string]any{
			{"produto_id": 10, "quantidade": 2},
		},
	})

	if res.Code != http.StatusCreated {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusCreated, res.Code, res.Body.String())
	}

	nota := decodeJSON[models.NotaFiscal](t, res)
	if nota.ID <= 0 || nota.Cliente != clientePadraoTeste || !nota.EstaAberta {
		t.Fatalf("nota criada invalida: %+v", nota)
	}
}

func TestListarNotasFiscaisRota(t *testing.T) {
	router := setupRouterFaturamento(t)

	_ = inserirNota(t, clientePadraoTeste, true, []models.ItemNota{{ProdutoID: 1, Quantidade: 1}})
	_ = inserirNota(t, "Cliente B", false, []models.ItemNota{{ProdutoID: 2, Quantidade: 2}})

	res := performJSONRequest(t, router, http.MethodGet, rotaNotasFiscaisTeste, nil)
	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	notas := decodeJSON[[]models.NotaFiscal](t, res)
	if len(notas) != 2 {
		t.Fatalf("quantidade inesperada de notas: %d", len(notas))
	}
}

func TestImprimirNotaRota(t *testing.T) {
	router := setupRouterFaturamento(t)
	nota := inserirNota(t, "Cliente Impressao", true, []models.ItemNota{{ProdutoID: 12, Quantidade: 3}})

	servidorEstoque := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/produtos/12/baixa" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set(headerContentTypeTeste, valorJSONTeste)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"mensagem":"ok"}`))
	}))
	defer servidorEstoque.Close()

	valorAnterior := os.Getenv("ESTOQUE_SERVICE_URL")
	t.Cleanup(func() {
		if valorAnterior == "" {
			_ = os.Unsetenv("ESTOQUE_SERVICE_URL")
			return
		}

		_ = os.Setenv("ESTOQUE_SERVICE_URL", valorAnterior)
	})

	if err := os.Setenv("ESTOQUE_SERVICE_URL", servidorEstoque.URL); err != nil {
		t.Fatalf("erro ao definir ESTOQUE_SERVICE_URL: %v", err)
	}

	res := performJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/notas-fiscais/%d/imprimir", nota.ID), map[string]any{})
	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	var notaAtualizada models.NotaFiscal
	if err := db.DB.First(&notaAtualizada, nota.ID).Error; err != nil {
		t.Fatalf("erro ao consultar nota atualizada: %v", err)
	}

	if notaAtualizada.EstaAberta {
		t.Fatalf("nota deveria estar fechada apos impressao")
	}
}

func TestCancelarNotaRota(t *testing.T) {
	router := setupRouterFaturamento(t)
	nota := inserirNota(t, "Cliente Cancelar", true, []models.ItemNota{{ProdutoID: 8, Quantidade: 1}})

	res := performJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/notas-fiscais/%d/cancelar", nota.ID), map[string]any{})
	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	var notaAtualizada models.NotaFiscal
	if err := db.DB.First(&notaAtualizada, nota.ID).Error; err != nil {
		t.Fatalf("erro ao consultar nota atualizada: %v", err)
	}

	if notaAtualizada.EstaAberta {
		t.Fatalf("nota deveria estar fechada apos cancelamento")
	}
}

func TestDeletarNotaRota(t *testing.T) {
	router := setupRouterFaturamento(t)
	nota := inserirNota(t, "Cliente Deletar", false, []models.ItemNota{{ProdutoID: 3, Quantidade: 2}})

	res := performJSONRequest(t, router, http.MethodDelete, fmt.Sprintf("/notas-fiscais/%d", nota.ID), nil)
	if res.Code != http.StatusOK {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusOK, res.Code, res.Body.String())
	}

	var notaRemovida models.NotaFiscal
	errNota := db.DB.First(&notaRemovida, nota.ID).Error
	if errNota == nil {
		t.Fatalf("nota nao foi removida")
	}

	var itens []models.ItemNota
	if err := db.DB.Where("nota_fiscal_id = ?", nota.ID).Find(&itens).Error; err != nil {
		t.Fatalf("erro ao consultar itens da nota: %v", err)
	}

	if len(itens) != 0 {
		t.Fatalf("itens da nota deveriam ter sido removidos, encontrados: %d", len(itens))
	}
}

func TestImprimirNotaFalhaIntegracaoMantemNotaAberta(t *testing.T) {
	router := setupRouterFaturamento(t)
	nota := inserirNota(t, "Cliente Falha Integracao", true, []models.ItemNota{{ProdutoID: 20, Quantidade: 2}})

	servidorEstoque := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set(headerContentTypeTeste, valorJSONTeste)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"erro":"Estoque indisponivel para impressao"}`))
	}))
	defer servidorEstoque.Close()

	valorAnterior := os.Getenv("ESTOQUE_SERVICE_URL")
	t.Cleanup(func() {
		if valorAnterior == "" {
			_ = os.Unsetenv("ESTOQUE_SERVICE_URL")
			return
		}

		_ = os.Setenv("ESTOQUE_SERVICE_URL", valorAnterior)
	})

	if err := os.Setenv("ESTOQUE_SERVICE_URL", servidorEstoque.URL); err != nil {
		t.Fatalf("erro ao definir ESTOQUE_SERVICE_URL: %v", err)
	}

	res := performJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/notas-fiscais/%d/imprimir", nota.ID), map[string]any{})
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusServiceUnavailable, res.Code, res.Body.String())
	}

	resposta := decodeJSON[map[string]any](t, res)
	if resposta["codigo"] != "BAIXA_ESTOQUE_FALHOU" {
		t.Fatalf(formatoCodigoErroInesperado, resposta)
	}

	var notaAtualizada models.NotaFiscal
	if err := db.DB.First(&notaAtualizada, nota.ID).Error; err != nil {
		t.Fatalf("erro ao consultar nota apos falha de integracao: %v", err)
	}

	if !notaAtualizada.EstaAberta {
		t.Fatalf("nota deveria permanecer aberta quando a integracao com estoque falha")
	}
}

func TestCancelarNotaFechadaRota(t *testing.T) {
	router := setupRouterFaturamento(t)
	nota := inserirNota(t, "Cliente Ja Fechado", false, []models.ItemNota{{ProdutoID: 30, Quantidade: 1}})

	res := performJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/notas-fiscais/%d/cancelar", nota.ID), map[string]any{})
	if res.Code != http.StatusConflict {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusConflict, res.Code, res.Body.String())
	}

	resposta := decodeJSON[map[string]any](t, res)
	if resposta["codigo"] != "NOTA_FECHADA" {
		t.Fatalf(formatoCodigoErroInesperado, resposta)
	}
}

func TestDeletarNotaAbertaRota(t *testing.T) {
	router := setupRouterFaturamento(t)
	nota := inserirNota(t, "Cliente Aberto", true, []models.ItemNota{{ProdutoID: 40, Quantidade: 1}})

	res := performJSONRequest(t, router, http.MethodDelete, fmt.Sprintf("/notas-fiscais/%d", nota.ID), nil)
	if res.Code != http.StatusConflict {
		t.Fatalf(formatoStatusInesperadoTeste, http.StatusConflict, res.Code, res.Body.String())
	}

	resposta := decodeJSON[map[string]any](t, res)
	if resposta["codigo"] != "NOTA_ABERTA" {
		t.Fatalf(formatoCodigoErroInesperado, resposta)
	}
}
