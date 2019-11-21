// Package connection establece una conexión con la base de datos
package connection

import (
	"database/sql"
	"fmt"
)

//ObtenerBaseDeDatos retorna la conexión a base de datos a utilizar
func ObtenerBaseDeDatos() (db *sql.DB, e error) {
	usuario := "root"
	pass := "1524863970"
	host := "tcp(127.0.0.1:3306)"
	nombre := "saitauditorias"
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", usuario, pass, host, nombre))
	if err != nil {
		return nil, err
	}
	return db, nil
}
