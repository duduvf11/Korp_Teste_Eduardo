package models

type NotaFiscal struct {
	ID       int        `json:"id"`
	Cliente  string     `json:"cliente"`
	Item     []ItemNota `json:"itens"`

}

type ItemNota struct {
	ID		  		int  `json:"id"`
	ProdutoID 		int	 `json:"produto_id"`
	NotaFiscalID 	int  `json:"nota_fiscal_id"`
	Quantidade 		int  `json:"quantidade"`
}