package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/doug-martin/goqu/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var dialect = goqu.Dialect("mysql")

//Usuario es una estructura de los datos de los usuarios en el sistema
type Usuario struct {
	ID     int    `json:"id"`
	Estado uint8  `json:"estado"`
	Email  string `json:"email"`
}

//Proyecto es una estructura de los datos de los proyectos en el sistema
type Proyecto struct {
	ID         int    `json:"id"`
	Nombre     string `json:"nombre"`
	Servidores []int  `json:"servidores"`
}

//Servidor es una estructura de los datos de los servidores en el sistema
type Servidor struct {
	ID               int       `json:"id"`
	DireccionPublica string    `json:"direccion_publica"`
	SistemaOperativo string    `json:"sistema_operativo"`
	Dominios         []Dominio `json:"dominios"`
}

//Dominio es una estructura de los datos de los dominios en el sistema
type Dominio struct {
	ID      int    `json:"id"`
	Dominio string `json:"dominio"`
}

//Auditoria es una estructura de los datos de las auditorías en el sistema
type Auditoria struct {
	ID             int    `json:"id"`
	Motivo         string `json:"motivo"`
	Comentario     string `json:"comentario"`
	Comandos       string `json:"comandos"`
	Fecha          string `json:"fecha"`
	Usuario        string `json:"usuario"`
	NombreProyecto string `json:"nombre_proyecto"`
	IPServidor     string `json:"ip_servidor"`
	IDUsuario      int    `json:"id_usuario"`
	IDProyecto     int    `json:"id_proyecto"`
	IDServidor     int    `json:"id_servidor"`
}

//Paginas es una estructura de los datos de las páginas para el sistema de paginado de los datos de las demás estructuras
type Paginas struct {
	NumeroPaginas  int `json:"numero_paginas"`
	TotalElementos int `json:"total_elementos"`
}

func obtenerBaseDeDatos() (db *sql.DB, e error) {
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

func getUsuarios(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getUsuario(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func createUsuario(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func updateUsuario(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getProyectos(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getProyectosPorServidores(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getProyecto(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func createProyecto(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func updateProyecto(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getServidores(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getServidoresPorProyecto(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getServidor(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func createServidor(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func updateServidor(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
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

func getAuditorias(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var auditoria Auditoria
	var selectSQL string
	var selectCountSQL string
	var pagina Paginas
	var page, _ = strconv.Atoi(c.DefaultQuery("pagina", "1"))
	offset := (uint)(page-1) * 10
	auditorias := []Auditoria{}
	t := time.Now()
	fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", t.Year(), t.Month(),
		t.Day(), t.Hour(), t.Minute(), t.Second())
	fechaLimitInf := c.DefaultQuery("limit_inf", "0000-00-00T00:00:00")
	fechaLimitSup := c.DefaultQuery("limit_sup", fecha)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if motivo := c.Query("motivo"); motivo != "" {
		selectSQL, _, _ = dialect.From("proyectos").Select(
			goqu.I("auditorias.id"),
			"motivo",
			"comentario",
			"comandos",
			"fecha",
			goqu.I("email").As("usuario"),
			goqu.I("nombre").As("nombre_proyecto"),
			goqu.I("direccion_publica").As("ip_servidor"),
		).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
		).Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
		).Join(
			goqu.T("auditorias"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("usuarios"),
			goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
		).Where(
			goqu.C("motivo").ILike(motivo+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).Order(goqu.C("fecha").Desc()).Offset(offset).Limit(10).ToSQL()
		selectCountSQL, _, _ = dialect.From("auditorias").Select(goqu.COUNT("*").As("total_elementos")).Where(
			goqu.C("motivo").ILike(motivo+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
		db.QueryRow(selectCountSQL).Scan(&pagina.TotalElementos)
		if (pagina.TotalElementos % 10) == 0 {
			pagina.NumeroPaginas = pagina.TotalElementos / 10
		} else {
			pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
		}
	} else if usuario := c.Query("usuario"); usuario != "" {
		selectSQL, _, _ = dialect.From("proyectos").Select(
			goqu.I("auditorias.id"),
			"motivo",
			"comentario",
			"comandos",
			"fecha",
			goqu.I("email").As("usuario"),
			goqu.I("nombre").As("nombre_proyecto"),
			goqu.I("direccion_publica").As("ip_servidor"),
		).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
		).Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
		).Join(
			goqu.T("auditorias"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("usuarios"),
			goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
		).Where(
			goqu.C("email").ILike(usuario+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).Order(goqu.C("fecha").Desc()).Offset(offset).Limit(10).ToSQL()
		selectCountSQL, _, _ = dialect.From("auditorias").Select(goqu.COUNT("*").As("total_elementos")).Join(
			goqu.T("usuarios"),
			goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
		).Where(
			goqu.C("email").ILike(usuario+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
		db.QueryRow(selectCountSQL).Scan(&pagina.TotalElementos)
		if (pagina.TotalElementos % 10) == 0 {
			pagina.NumeroPaginas = pagina.TotalElementos / 10
		} else {
			pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
		}
	} else if nombreProyecto := c.Query("nombre_proyecto"); nombreProyecto != "" {
		selectSQL, _, _ = dialect.From("proyectos").Select(
			goqu.I("auditorias.id"),
			"motivo",
			"comentario",
			"comandos",
			"fecha",
			goqu.I("email").As("usuario"),
			goqu.I("nombre").As("nombre_proyecto"),
			goqu.I("direccion_publica").As("ip_servidor"),
		).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
		).Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
		).Join(
			goqu.T("auditorias"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("usuarios"),
			goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
		).Where(
			goqu.C("nombre").ILike("%"+nombreProyecto+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).Order(goqu.C("fecha").Desc()).Offset(offset).Limit(10).ToSQL()
		selectCountSQL, _, _ = dialect.From("auditorias").Select(goqu.COUNT("*").As("total_elementos")).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("proyectos"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
		).Where(
			goqu.C("nombre").ILike("%"+nombreProyecto+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
		db.QueryRow(selectCountSQL).Scan(&pagina.TotalElementos)
		if (pagina.TotalElementos % 10) == 0 {
			pagina.NumeroPaginas = pagina.TotalElementos / 10
		} else {
			pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
		}
	} else if servidorIP := c.Query("ip_servidor"); servidorIP != "" {
		selectSQL, _, _ = dialect.From("proyectos").Select(
			goqu.I("auditorias.id"),
			"motivo",
			"comentario",
			"comandos",
			"fecha",
			goqu.I("email").As("usuario"),
			goqu.I("nombre").As("nombre_proyecto"),
			goqu.I("direccion_publica").As("ip_servidor"),
		).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
		).Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
		).Join(
			goqu.T("auditorias"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("usuarios"),
			goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
		).Where(
			goqu.C("direccion_publica").ILike(servidorIP+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).Order(goqu.C("fecha").Desc()).Offset(offset).Limit(10).ToSQL()
		selectCountSQL, _, _ = dialect.From("auditorias").Select(goqu.COUNT("*").As("total_elementos")).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
		).Where(
			goqu.C("direccion_publica").ILike(servidorIP+"%"),
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
		db.QueryRow(selectCountSQL).Scan(&pagina.TotalElementos)
		if (pagina.TotalElementos % 10) == 0 {
			pagina.NumeroPaginas = pagina.TotalElementos / 10
		} else {
			pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
		}
	} else {
		selectSQL, _, _ = dialect.From("proyectos").Select(
			goqu.I("auditorias.id"),
			"motivo",
			"comentario",
			"comandos",
			"fecha",
			goqu.I("email").As("usuario"),
			goqu.I("nombre").As("nombre_proyecto"),
			goqu.I("direccion_publica").As("ip_servidor"),
		).Join(
			goqu.T("proyectos_servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
		).Join(
			goqu.T("servidores"),
			goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
		).Join(
			goqu.T("auditorias"),
			goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
		).Join(
			goqu.T("usuarios"),
			goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
		).Where(
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).Order(goqu.C("fecha").Desc()).Offset(offset).Limit(10).ToSQL()
		selectCountSQL, _, _ = dialect.From("auditorias").Select(goqu.COUNT("*").As("total_elementos")).Where(
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
		db.QueryRow(selectCountSQL).Scan(&pagina.TotalElementos)
		if (pagina.TotalElementos % 10) == 0 {
			pagina.NumeroPaginas = pagina.TotalElementos / 10
		} else {
			pagina.NumeroPaginas = (pagina.TotalElementos / 10) + 1
		}
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&auditoria.ID, &auditoria.Motivo, &auditoria.Comentario, &auditoria.Comandos, &auditoria.Fecha, &auditoria.Usuario, &auditoria.NombreProyecto, &auditoria.IPServidor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		auditorias = append(auditorias, auditoria)
	}
	c.JSON(http.StatusOK, gin.H{
		"auditorias": auditorias,
		"pagina":     pagina,
	})
}

func getAuditoria(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var auditoria Auditoria
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ := dialect.From("proyectos").Select(
		goqu.I("auditorias.id"),
		"motivo",
		"comentario",
		"comandos",
		"fecha",
		goqu.I("email").As("usuario"),
		goqu.I("nombre").As("nombre_proyecto"),
		goqu.I("direccion_publica").As("ip_servidor"),
	).Join(
		goqu.T("proyectos_servidores"),
		goqu.On(goqu.Ex{"proyectos_servidores.id_proyecto": goqu.I("proyectos.id")}),
	).Join(
		goqu.T("servidores"),
		goqu.On(goqu.Ex{"proyectos_servidores.id_servidor": goqu.I("servidores.id")}),
	).Join(
		goqu.T("auditorias"),
		goqu.On(goqu.Ex{"proyectos_servidores.id": goqu.I("auditorias.id_proyecto_servidor")}),
	).Join(
		goqu.T("usuarios"),
		goqu.On(goqu.Ex{"auditorias.id_usuario": goqu.I("usuarios.id")}),
	).Where(goqu.Ex{"auditorias.id": c.Param("id")}).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&auditoria.ID, &auditoria.Motivo, &auditoria.Comentario,
		&auditoria.Comandos, &auditoria.Fecha, &auditoria.Usuario, &auditoria.NombreProyecto, &auditoria.IPServidor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, auditoria)
}

func createAuditoria(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var auditoria Auditoria
	var idProyectoServidor int
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
	err = json.Unmarshal(body, &auditoria)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	selectSQL, _, _ := dialect.From("proyectos_servidores").Select("id").Where(
		goqu.Ex{"id_proyecto": auditoria.IDProyecto},
		goqu.Ex{"id_servidor": auditoria.IDServidor},
	).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&idProyectoServidor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	insertSQL, _, _ := dialect.Insert("auditorias").Rows(
		goqu.Record{"motivo": auditoria.Motivo, "comentario": auditoria.Comentario,
			"comandos": auditoria.Comandos, "id_usuario": auditoria.IDUsuario, "id_proyecto_servidor": idProyectoServidor},
	).ToSQL()
	result, err := db.Exec(insertSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	auditoria.ID = int(id)
	c.JSON(http.StatusCreated, auditoria)
}

func main() {
	db, err := obtenerBaseDeDatos()
	if err != nil {
		fmt.Printf("Error obteniendo base de datos: %v", err)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Printf("Error conectando: %v", err)
		return
	}
	fmt.Println("Conectado correctamente")

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://192.168.1.136:8080", "http://localhost:8080"}

	r.Use(cors.New(config))
	r.GET("/usuarios", getUsuarios)
	r.GET("/usuarios/:id", getUsuario)
	r.POST("/usuarios", createUsuario)
	r.PUT("/usuarios/:id", updateUsuario)
	r.GET("/proyectos", getProyectos)
	r.GET("/servidores/:id/proyectos", getProyectosPorServidores)
	r.GET("/proyectos/:id", getProyecto)
	r.POST("/proyectos", createProyecto)
	r.PUT("/proyectos/:id", updateProyecto)
	r.GET("/servidores", getServidores)
	r.GET("/proyectos/:id/servidores", getServidoresPorProyecto)
	r.GET("/servidores/:id", getServidor)
	r.POST("/servidores", createServidor)
	r.PUT("/servidores/:id", updateServidor)
	r.GET("/auditorias", getAuditorias)
	r.GET("/auditorias/:id", getAuditoria)
	r.POST("/auditorias", createAuditoria)
	r.Run(":3000")
}
