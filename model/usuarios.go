// Package model modela las estructuras y funciones para acceder a la base de datos
package model

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flix14/audit/connection"

	"github.com/doug-martin/goqu"
	"github.com/gin-gonic/gin"
)

//Usuario es una estructura de los datos de los usuarios en el sistema
type Usuario struct {
	ID     int    `json:"id"`
	Estado uint8  `json:"estado"`
	Email  string `json:"email"`
}

//GetUsuarios obtiene los usuarios almacenados en la base de datos
func GetUsuarios(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var usuario Usuario
	var selectSQL string
	var selectPaginaSQL string
	var pagina Paginas
	var page, _ = strconv.Atoi(c.DefaultQuery("pagina", "0"))
	offset := (uint)(page-1) * 10
	usuarios := []Usuario{}
	estado1 := c.DefaultQuery("estado", "0")
	estado2 := c.DefaultQuery("estado", "1")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if email := c.Query("email"); email != "" {
		if page <= 0 {
			selectSQL, _, _ = dialect.From("usuarios").Where(goqu.And(
				goqu.C("email").ILike(email+"%"),
				goqu.Ex{"estado": []string{estado1, estado2}},
			)).Order(goqu.C("id").Desc()).Limit(10).ToSQL()
		} else {
			selectSQL, _, _ = dialect.From("usuarios").Where(goqu.And(
				goqu.C("email").ILike(email+"%"),
				goqu.Ex{"estado": []string{estado1, estado2}},
			)).Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		}
		selectPaginaSQL, _, _ = dialect.From("usuarios").Select(goqu.COUNT("*").As("total_elementos")).Where(goqu.And(
			goqu.C("email").ILike(email+"%"),
			goqu.Ex{"estado": []string{estado1, estado2}},
		)).ToSQL()
	} else if page > 0 {
		selectSQL, _, _ = dialect.From("usuarios").Where(goqu.Ex{
			"estado": []string{estado1, estado2},
		}).Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		selectPaginaSQL, _, _ = dialect.From("usuarios").Select(goqu.COUNT("*").As("total_elementos")).Where(goqu.Ex{
			"estado": []string{estado1, estado2},
		}).ToSQL()
	} else {
		selectSQL, _, _ = dialect.From("usuarios").Where(goqu.Ex{
			"estado": []string{estado1, estado2},
		}).Order(goqu.C("id").Desc()).ToSQL()
		selectPaginaSQL, _, _ = dialect.From("usuarios").Select(goqu.COUNT("*").As("total_elementos")).Where(goqu.Ex{
			"estado": []string{estado1, estado2},
		}).ToSQL()
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&usuario.ID, &usuario.Estado, &usuario.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		usuarios = append(usuarios, usuario)
	}
	db.QueryRow(selectPaginaSQL).Scan(&pagina.TotalElementos)
	if (pagina.TotalElementos % 10) == 0 {
		pagina.NumeroPaginas = pagina.TotalElementos / 10
	} else {
		pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
	}
	c.JSON(http.StatusOK, gin.H{
		"usuarios": usuarios,
		"pagina":   pagina,
	})
}

//GetUsuario obtiene un usuario de la base de datos por ID
func GetUsuario(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var usuario Usuario
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ := dialect.From("usuarios").Where(goqu.Ex{
		"id": c.Param("id"),
	}).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&usuario.ID, &usuario.Estado, &usuario.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, usuario)
}

//CreateUsuario registra a nuevo usuario en la base de datos
func CreateUsuario(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var usuario Usuario
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = json.Unmarshal(body, &usuario)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	insertSQL, _, _ := dialect.Insert("usuarios").Rows(
		goqu.Record{"email": usuario.Email, "estado": usuario.Estado},
	).ToSQL()
	result, err := db.Exec(insertSQL)
	if err != nil {
		selectSQL, _, _ := dialect.From("usuarios").Select("id", "estado").Where(goqu.Ex{
			"email": usuario.Email,
		}).ToSQL()
		fila := db.QueryRow(selectSQL)
		err = fila.Scan(&usuario.ID, &usuario.Estado)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.Header("Location", "/usuarios/"+strconv.Itoa(usuario.ID))
			c.JSON(http.StatusConflict, usuario)
		}
		return
	}
	id, _ := result.LastInsertId()
	usuario.ID = int(id)
	c.JSON(http.StatusCreated, usuario)
}

//UpdateUsuario actualiza los datos de un usuario en la base de datos
func UpdateUsuario(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var usuario Usuario
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = json.Unmarshal(body, &usuario)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	updateSQL, _, _ := dialect.Update("usuarios").Set(
		goqu.Record{"estado": usuario.Estado, "email": usuario.Email},
	).Where(
		goqu.Ex{"id": c.Param("id")},
	).ToSQL()
	result, err := db.Exec(updateSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rows affected: 0"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	usuario.ID = int(id)
	c.JSON(http.StatusOK, usuario)
}
