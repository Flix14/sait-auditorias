// Package model modela las estructuras y funciones para acceder a la base de datos
package model

import (
	"github.com/doug-martin/goqu"
)

var dialect = goqu.Dialect("mysql")

//Paginas es una estructura de los datos de las páginas para el sistema de paginado de los datos de las demás estructuras
type Paginas struct {
	NumeroPaginas  int `json:"numero_paginas"`
	TotalElementos int `json:"total_elementos"`
}
