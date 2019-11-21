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

//Servidor es una estructura de los datos de los servidores en el sistema
type Servidor struct {
	ID               int       `json:"id"`
	DireccionPublica string    `json:"direccion_publica"`
	SistemaOperativo string    `json:"sistema_operativo"`
	Dominios         []Dominio `json:"dominios"`
}

//Dominio es una estructura de los datos de los dominios de los servidores
type Dominio struct {
	ID      int    `json:"id"`
	Dominio string `json:"dominio"`
}

//GetServidores obtiene los servidores almacenados en la base de datos
func GetServidores(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var servidor Servidor
	var dominio Dominio
	var selectSQL string
	var selectPaginaSQL string
	var pagina Paginas
	var page, _ = strconv.Atoi(c.DefaultQuery("pagina", "0"))
	offset := (uint)(page-1) * 10
	servidores := []Servidor{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if direccionPublica := c.Query("direccion_publica"); direccionPublica != "" {
		if page <= 0 {
			selectSQL, _, _ = dialect.From("servidores").Where(
				goqu.C("direccion_publica").ILike(direccionPublica + "%")).Order(goqu.C("id").Desc()).Limit(10).ToSQL()
		} else {
			selectSQL, _, _ = dialect.From("servidores").Where(
				goqu.C("direccion_publica").ILike(direccionPublica + "%")).Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		}
		selectPaginaSQL, _, _ = dialect.From("servidores").Select(goqu.COUNT("*").As("total_elementos")).Where(
			goqu.C("direccion_publica").ILike(direccionPublica + "%")).ToSQL()
	} else if dominio := c.Query("dominio"); dominio != "" {
		if offset <= 0 {
			selectSQL, _, _ = dialect.From("servidores").Select(
				goqu.I("servidores.id"), "direccion_publica", "sistema_operativo").Distinct().Join(
				goqu.T("dominios"),
				goqu.On(goqu.Ex{"servidores.id": goqu.I("dominios.id_servidor")}),
			).Where(
				goqu.C("dominio").ILike("%" + dominio + "%")).Order(goqu.C("id").Desc()).Limit(10).ToSQL()
		} else {
			selectSQL, _, _ = dialect.From("servidores").Select(
				goqu.I("servidores.id"), "direccion_publica", "sistema_operativo").Distinct().Join(
				goqu.T("dominios"),
				goqu.On(goqu.Ex{"servidores.id": goqu.I("dominios.id_servidor")}),
			).Where(
				goqu.C("dominio").ILike("%" + dominio + "%")).Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		}
		selectPaginaSQL, _, _ = dialect.From("servidores").Select(
			goqu.COUNT(goqu.DISTINCT(goqu.I("servidores.id"))).As("total_elementos"),
		).Join(
			goqu.T("dominios"),
			goqu.On(goqu.Ex{"servidores.id": goqu.I("dominios.id_servidor")}),
		).Where(
			goqu.C("dominio").ILike("%" + dominio + "%")).ToSQL()
	} else if page > 0 {
		selectSQL, _, _ = dialect.From("servidores").Order(goqu.C("id").Desc()).Offset(offset).Limit(10).ToSQL()
		selectPaginaSQL, _, _ = dialect.From("servidores").Select(goqu.COUNT("*").As("total_elementos")).ToSQL()
	} else {
		selectSQL, _, _ = dialect.From("servidores").Order(goqu.C("id").Desc()).ToSQL()
		selectPaginaSQL, _, _ = dialect.From("servidores").Select(goqu.COUNT("*").As("total_elementos")).ToSQL()
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		servidores = append(servidores, servidor)
	}
	for index := range servidores {
		serv := &servidores[index]
		var listaDominios []Dominio
		selectSQL, _, _ = dialect.From("dominios").Select("dominios.id", "dominio").Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"servidores.id": goqu.I("dominios.id_servidor")}),
		).Where(goqu.Ex{
			"servidores.id": serv.ID,
		}).ToSQL()
		filas, err = db.Query(selectSQL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer filas.Close()
		for filas.Next() {
			err = filas.Scan(&dominio.ID, &dominio.Dominio)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			listaDominios = append(listaDominios, dominio)
		}
		serv.Dominios = listaDominios
	}
	db.QueryRow(selectPaginaSQL).Scan(&pagina.TotalElementos)
	if (pagina.TotalElementos % 10) == 0 {
		pagina.NumeroPaginas = pagina.TotalElementos / 10
	} else {
		pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
	}
	c.JSON(http.StatusOK, gin.H{
		"servidores": servidores,
		"pagina":     pagina,
	})
}

//GetServidoresPorProyecto obtiene los servdiores pertenecientes al proyecto seleccionado por ID
func GetServidoresPorProyecto(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var servidor Servidor
	var selectSQL string
	servidores := []Servidor{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ = dialect.From("servidores").Where(goqu.Ex{
		"servidores.id": dialect.From("proyectos_servidores").Select(
			"id_servidor",
		).Where(goqu.Ex{
			"id_proyecto": c.Param("id"),
		}),
	}).ToSQL()
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		servidores = append(servidores, servidor)
	}
	c.JSON(http.StatusOK, servidores)
}

//GetServidor obtiene un servidor de la base de datos por ID
func GetServidor(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var servidor Servidor
	var dominio Dominio
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ := dialect.From("servidores").Where(goqu.Ex{
		"id": c.Param("id"),
	}).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var listaDominios []Dominio
	selectSQL, _, _ = dialect.From("dominios").Select("dominios.id", "dominio").Join(
		goqu.T("servidores"),
		goqu.On(goqu.Ex{"servidores.id": goqu.I("dominios.id_servidor")}),
	).Where(goqu.Ex{
		"servidores.id": servidor.ID,
	}).ToSQL()
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&dominio.ID, &dominio.Dominio)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		listaDominios = append(listaDominios, dominio)
	}
	servidor.Dominios = listaDominios

	c.JSON(http.StatusOK, servidor)
}

//CreateServidor registra un nuevo servidor en la base de datos
func CreateServidor(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var servidor Servidor
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
	err = json.Unmarshal(body, &servidor)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	insertSQL, _, _ = dialect.Insert("servidores").Rows(
		goqu.Record{"direccion_publica": servidor.DireccionPublica, "sistema_operativo": servidor.SistemaOperativo},
	).ToSQL()
	result, err := db.Exec(insertSQL)
	if err != nil {
		selectSQL, _, _ := dialect.From("servidores").Select("id", "sistema_operativo").Where(goqu.Ex{
			"direccion_publica": servidor.DireccionPublica,
		}).ToSQL()
		fila := db.QueryRow(selectSQL)
		err = fila.Scan(&servidor.ID, &servidor.SistemaOperativo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.Header("Location", "/servidores/"+strconv.Itoa(servidor.ID))
			c.JSON(http.StatusConflict, servidor)
		}
		return
	}
	id, _ := result.LastInsertId()
	servidor.ID = int(id)

	for _, serv := range servidor.Dominios {
		insertSQL, _, _ = dialect.Insert("dominios").Rows(
			goqu.Record{"id_servidor": servidor.ID, "dominio": serv.Dominio},
		).ToSQL()
		_, err = db.Exec(insertSQL)
		if err != nil {
			deleteSQL, _, _ := dialect.Delete("dominios").Where(goqu.Ex{
				"id_servidor": servidor.ID,
			}).ToSQL()
			db.Exec(deleteSQL)
			deleteSQL, _, _ = dialect.Delete("servidores").Where(goqu.Ex{
				"id": servidor.ID,
			}).ToSQL()
			db.Exec(deleteSQL)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	}

	c.JSON(http.StatusCreated, servidor)
}

//UpdateServidor actualiza los datos de un servidor en la base de datos
func UpdateServidor(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
	var servidor Servidor
	var errors []string
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
	err = json.Unmarshal(body, &servidor)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	updateSQL, _, _ := dialect.Update("servidores").Set(
		goqu.Record{"direccion_publica": servidor.DireccionPublica, "sistema_operativo": servidor.SistemaOperativo},
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
	servidor.ID = int(id)

	deleteSQL, _, _ := dialect.Delete("dominios").Where(goqu.Ex{
		"id_servidor": servidor.ID,
	}).ToSQL()
	db.Exec(deleteSQL)

	for _, serv := range servidor.Dominios {
		insertSQL, _, _ := dialect.Insert("dominios").Rows(
			goqu.Record{"id_servidor": servidor.ID, "dominio": serv.Dominio},
		).ToSQL()
		_, err = db.Exec(insertSQL)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"errors": errors})
		return
	}

	c.JSON(http.StatusOK, servidor)
}
