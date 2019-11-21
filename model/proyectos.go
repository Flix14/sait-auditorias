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

//Proyecto es una estructura de los datos de los proyectos en el sistema
type Proyecto struct {
	ID         int    `json:"id"`
	Nombre     string `json:"nombre"`
	Servidores []int  `json:"servidores"`
}

//GetProyectos obtiene los proyectos almacenados en la base de datos
func GetProyectos(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var proyecto Proyecto
	var selectSQL string
	var selectPaginaSQL string
	var pagina Paginas
	var page, _ = strconv.Atoi(c.DefaultQuery("pagina", "0"))
	offset := (uint)(page-1) * 10
	proyectos := []Proyecto{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if nombre := c.Query("nombre"); nombre != "" {
		if page <= 0 {
			selectSQL, _, _ = dialect.From("proyectos").Where(
				goqu.C("nombre").ILike(nombre + "%"),
			).Order(goqu.C("id").Desc()).Limit(10).ToSQL()
		} else {
			selectSQL, _, _ = dialect.From("proyectos").Where(
				goqu.C("nombre").ILike(nombre + "%"),
			).Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		}
		selectPaginaSQL, _, _ = dialect.From("proyectos").Select(goqu.COUNT("*").As("total_elementos")).Where(
			goqu.C("nombre").ILike(nombre + "%"),
		).ToSQL()
	} else if page > 0 {
		selectSQL, _, _ = dialect.From("proyectos").Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		selectPaginaSQL, _, _ = dialect.From("proyectos").Select(goqu.COUNT("*").As("total_elementos")).ToSQL()
	} else {
		selectSQL, _, _ = dialect.From("proyectos").Order(goqu.C("id").Desc()).ToSQL()
		selectPaginaSQL, _, _ = dialect.From("proyectos").Select(goqu.COUNT("*").As("total_elementos")).ToSQL()
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&proyecto.ID, &proyecto.Nombre)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		proyectos = append(proyectos, proyecto)
	}

	db.QueryRow(selectPaginaSQL).Scan(&pagina.TotalElementos)
	if (pagina.TotalElementos % 10) == 0 {
		pagina.NumeroPaginas = pagina.TotalElementos / 10
	} else {
		pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
	}
	c.JSON(http.StatusOK, gin.H{
		"proyectos": proyectos,
		"pagina":    pagina,
	})
}

//GetProyectosPorServidores obtiene los proyectos del servidor seleccionado por ID
func GetProyectosPorServidores(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var proyecto Proyecto
	var selectSQL string
	proyectos := []Proyecto{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ = dialect.From("proyectos").Where(goqu.Ex{
		"proyectos.id": dialect.From("proyectos_servidores").Select(
			"id_proyecto",
		).Where(goqu.Ex{
			"id_servidor": c.Param("id"),
		}),
	}).ToSQL()
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&proyecto.ID, &proyecto.Nombre)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		proyectos = append(proyectos, proyecto)
	}
	c.JSON(http.StatusOK, proyectos)
}

//GetProyecto obtiene un proyecto de la base de datos por ID
func GetProyecto(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var proyecto Proyecto
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ := dialect.From("proyectos").Where(goqu.Ex{
		"id": c.Param("id"),
	}).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&proyecto.ID, &proyecto.Nombre)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proyecto)
}

//CreateProyecto registra un nuevo proyecto en la base de datos
func CreateProyecto(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var proyecto Proyecto
	var insertSQL string
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
	err = json.Unmarshal(body, &proyecto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	insertSQL, _, _ = dialect.Insert("proyectos").Rows(
		goqu.Record{"nombre": proyecto.Nombre},
	).ToSQL()
	result, err := db.Exec(insertSQL)
	if err != nil {
		selectSQL, _, _ := dialect.From("proyectos").Select("id").Where(goqu.Ex{
			"nombre": proyecto.Nombre,
		}).ToSQL()
		fila := db.QueryRow(selectSQL)
		err = fila.Scan(&proyecto.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.Header("Location", "/proyectos/"+strconv.Itoa(proyecto.ID))
			c.JSON(http.StatusConflict, proyecto)
		}
		return
	}
	id, _ := result.LastInsertId()
	for _, servidorID := range proyecto.Servidores {
		insertSQL, _, _ = dialect.Insert("proyectos_servidores").Rows(
			goqu.Record{"id_proyecto": id, "id_servidor": servidorID},
		).ToSQL()
		db.Exec(insertSQL)
	}
	proyecto.ID = int(id)
	c.JSON(http.StatusCreated, proyecto)
}

//UpdateProyecto actualiza los datos de un proyecto en la base de datos
func UpdateProyecto(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var proyecto Proyecto
	var insertSQL string
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
	err = json.Unmarshal(body, &proyecto)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	updateSQL, _, _ := dialect.Update("proyectos").Set(
		goqu.Record{"nombre": proyecto.Nombre},
	).Where(
		goqu.Ex{"id": c.Param("id")},
	).ToSQL()
	_, err = db.Exec(updateSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	proyecto.ID = int(id)

	deleteSQL, _, _ := dialect.Delete("proyectos_servidores").Where(goqu.Ex{
		"id_proyecto": proyecto.ID,
	}).ToSQL()
	db.Exec(deleteSQL)

	for _, servidorID := range proyecto.Servidores {
		insertSQL, _, _ = dialect.Insert("proyectos_servidores").Rows(
			goqu.Record{"id_proyecto": proyecto.ID, "id_servidor": servidorID},
		).ToSQL()
		db.Exec(insertSQL)
	}

	c.JSON(http.StatusOK, proyecto)
}
