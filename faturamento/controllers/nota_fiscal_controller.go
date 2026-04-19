package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/db"
	"github.com/duduvf11/Korp_Teste_Eduardo/faturamento/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type erroIntegracaoEstoque struct {
	statusCode int
	codigo     string
	mensagem   string
	detalhes   any
}

func responderErro(c *gin.Context, status int, codigo, mensagem string, detalhes any) {
	resposta := gin.H{
		"codigo": codigo,
		"erro":   mensagem,
	}

	if detalhes != nil {
		resposta["detalhes"] = detalhes
	}

	c.JSON(status, resposta)
}

func obterURLServicoEstoque() string {
	url := strings.TrimSpace(os.Getenv("ESTOQUE_SERVICE_URL"))
	if url == "" {
		return "http://localhost:8080"
	}

	return strings.TrimRight(url, "/")
}

func extrairMensagemErroDoEstoque(corpo []byte) string {
	if len(corpo) == 0 {
		return ""
	}

	var resposta map[string]any
	if err := json.Unmarshal(corpo, &resposta); err == nil {
		for _, chave := range []string{"erro", "error", "mensagem"} {
			valor, existe := resposta[chave]
			if !existe {
				continue
			}

			mensagem, ok := valor.(string)
			if ok && strings.TrimSpace(mensagem) != "" {
				return mensagem
			}
		}
	}

	return strings.TrimSpace(string(corpo))
}

func solicitarBaixaEstoque(cliente *http.Client, urlBase string, item models.ItemNota) *erroIntegracaoEstoque {
	payload := map[string]int{"quantidade": item.Quantidade}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &erroIntegracaoEstoque{
			statusCode: http.StatusInternalServerError,
			codigo:     "ERRO_SERIALIZACAO_BAIXA",
			mensagem:   "Nao foi possivel preparar a requisicao de baixa no estoque.",
			detalhes:   err.Error(),
		}
	}

	urlEstoque := fmt.Sprintf("%s/produtos/%d/baixa", urlBase, item.ProdutoID)
	req, err := http.NewRequest(http.MethodPost, urlEstoque, bytes.NewBuffer(jsonData))
	if err != nil {
		return &erroIntegracaoEstoque{
			statusCode: http.StatusInternalServerError,
			codigo:     "ERRO_REQUISICAO_BAIXA",
			mensagem:   "Nao foi possivel montar a requisicao para o estoque.",
			detalhes:   err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cliente.Do(req)
	if err != nil {
		return &erroIntegracaoEstoque{
			statusCode: http.StatusServiceUnavailable,
			codigo:     "ESTOQUE_INDISPONIVEL",
			mensagem:   "O servico de estoque esta temporariamente indisponivel.",
			detalhes:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	corpo, _ := io.ReadAll(resp.Body)
	mensagem := extrairMensagemErroDoEstoque(corpo)
	if mensagem == "" {
		mensagem = fmt.Sprintf("Nao foi possivel dar baixa no produto ID %d.", item.ProdutoID)
	}

	return &erroIntegracaoEstoque{
		statusCode: resp.StatusCode,
		codigo:     "BAIXA_ESTOQUE_FALHOU",
		mensagem:   mensagem,
		detalhes: gin.H{
			"produto_id":       item.ProdutoID,
			"status_estoque":   resp.StatusCode,
			"quantidade":       item.Quantidade,
			"endpoint_estoque": urlEstoque,
		},
	}
}

func validarItensNota(itens []models.ItemNota) error {
	if len(itens) == 0 {
		return errors.New("a nota fiscal precisa ter ao menos um item")
	}

	for _, item := range itens {
		if item.ProdutoID <= 0 {
			return errors.New("todos os itens devem possuir produto_id valido")
		}

		if item.Quantidade <= 0 {
			return errors.New("todos os itens devem possuir quantidade maior que zero")
		}
	}

	return nil
}

func CriarNotaFiscal(c *gin.Context) {
	var nota models.NotaFiscal
	if err := c.ShouldBindJSON(&nota); err != nil {
		responderErro(c, http.StatusBadRequest, "DADOS_INVALIDOS", "Dados invalidos para criar nota fiscal.", err.Error())
		return
	}

	if strings.TrimSpace(nota.Cliente) == "" {
		responderErro(c, http.StatusBadRequest, "CLIENTE_OBRIGATORIO", "O campo cliente e obrigatorio.", nil)
		return
	}

	if err := validarItensNota(nota.Item); err != nil {
		responderErro(c, http.StatusBadRequest, "ITENS_INVALIDOS", err.Error(), nil)
		return
	}

	nota.EstaAberta = true

	if err := db.DB.Create(&nota).Error; err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel salvar a nota fiscal.", err.Error())
		return
	}

	c.JSON(http.StatusCreated, nota)

}

func ListarNotasFiscais(c *gin.Context) {
	var notas []models.NotaFiscal

	if err := db.DB.Preload("Item").Find(&notas).Error; err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel listar as notas fiscais.", err.Error())
		return
	}

	c.JSON(http.StatusOK, notas)
}

func ImprimirNota(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		responderErro(c, http.StatusBadRequest, "ID_INVALIDO", "ID da nota fiscal invalido.", nil)
		return
	}

	var nota models.NotaFiscal

	err = db.DB.Preload("Item").First(&nota, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responderErro(c, http.StatusNotFound, "NOTA_NAO_ENCONTRADA", "Nota fiscal nao encontrada.", nil)
		return
	}

	if err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel consultar a nota fiscal.", err.Error())
		return
	}

	if !nota.EstaAberta {
		responderErro(c, http.StatusConflict, "NOTA_FECHADA", "Esta nota ja esta fechada e nao pode ser impressa novamente.", gin.H{"nota_id": nota.ID})
		return
	}

	if err := validarItensNota(nota.Item); err != nil {
		responderErro(c, http.StatusBadRequest, "ITENS_INVALIDOS", err.Error(), gin.H{"nota_id": nota.ID})
		return
	}

	clienteHTTP := &http.Client{Timeout: 5 * time.Second}
	urlBaseEstoque := obterURLServicoEstoque()

	for _, item := range nota.Item {
		errIntegracao := solicitarBaixaEstoque(clienteHTTP, urlBaseEstoque, item)
		if errIntegracao != nil {
			responderErro(c, errIntegracao.statusCode, errIntegracao.codigo, errIntegracao.mensagem, gin.H{
				"nota_id":  nota.ID,
				"item":     item,
				"detalhes": errIntegracao.detalhes,
			})
			return
		}
	}

	nota.EstaAberta = false
	if err := db.DB.Save(&nota).Error; err != nil {
		responderErro(c, http.StatusInternalServerError, "ERRO_BANCO", "Nao foi possivel atualizar o status da nota fiscal.", gin.H{
			"nota_id":  nota.ID,
			"problema": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem": "Nota fiscal impressa e estoque atualizado com sucesso.",
		"nota":     nota,
	})
}
