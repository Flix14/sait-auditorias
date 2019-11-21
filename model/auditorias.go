// Package model modela las estructuras y funciones para acceder a la base de datos
package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flix14/audit/connection"

	"github.com/doug-martin/goqu"
	"github.com/gin-gonic/gin"
)

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

//GetAuditorias obtiene las auditorías almacenadas en la base de datos
func GetAuditorias(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
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

//GetAuditoria obtiene una auditoría de la base de datos por ID
func GetAuditoria(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
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

//CreateAuditoria registra una nueva auditoría en la base de datos
func CreateAuditoria(c *gin.Context) {
	db, err := connection.ObtenerBaseDeDatos()
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
