package middlewares

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-Id"

func RequestContextLogger(servico string) gin.HandlerFunc {
	return func(c *gin.Context) {
		inicio := time.Now()

		requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = gerarRequestID()
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)

		c.Next()

		duracaoMs := time.Since(inicio).Milliseconds()
		rota := c.FullPath()
		if rota == "" {
			rota = c.Request.URL.Path
		}

		status := c.Writer.Status()
		codigoErro := contextoComoString(c, "erro_codigo")
		mensagemErro := contextoComoString(c, "erro_mensagem")

		if status >= http.StatusBadRequest {
			log.Printf(
				"service=%s level=error request_id=%s method=%s path=%s status=%d duration_ms=%d client_ip=%s error_code=%s error_message=%q",
				servico,
				requestID,
				c.Request.Method,
				rota,
				status,
				duracaoMs,
				c.ClientIP(),
				codigoErro,
				mensagemErro,
			)
			return
		}

		log.Printf(
			"service=%s level=info request_id=%s method=%s path=%s status=%d duration_ms=%d client_ip=%s",
			servico,
			requestID,
			c.Request.Method,
			rota,
			status,
			duracaoMs,
			c.ClientIP(),
		)
	}
}

func gerarRequestID() string {
	return "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func contextoComoString(c *gin.Context, chave string) string {
	valor, existe := c.Get(chave)
	if !existe {
		return ""
	}

	texto, ok := valor.(string)
	if !ok {
		return ""
	}

	return texto
}
