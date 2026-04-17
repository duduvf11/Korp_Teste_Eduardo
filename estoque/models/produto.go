package models

type Produto struct {
	Codigo 			int 		`json:"codigo"`
	Descricao 		string 		`json:"descricao"`
	Saldo 			int 		`json:"saldo"`
	Preco 			float64 	`json:"preco"`
}