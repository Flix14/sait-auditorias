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
	"github.com/gin-gonic/gin"

	_ "github.com/doug-martin/goqu/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var dialect = goqu.Dialect("mysql")

//Usuario description
type Usuario struct {
	ID     int    `json:"id"`
	Estado uint8  `json:"estado"`
	Email  string `json:"email"`
}

//Proyecto description Debe tener lista de servidores en los que se encuentra el proyecto
type Proyecto struct {
	ID         int    `json:"id"`
	Nombre     string `json:"nombre"`
	Servidores []int  `json:"servidores"`
	//Traer (get) de proyectos-servidores lista de servidores que pertenecen al proyecto (los servidores no son unique)
	//En Proyecto es donde se insertan las relaciones de la tabla proyectos-servidores
	//Ver la opción de agregar/eliminar servidor en un proyecto existente
	//Propuesta de uri /proyectos/:id/servidores (GET) y /proyectos/:id/servidores/:id (DELETE)
}

//Servidor description
type Servidor struct {
	ID               int    `json:"id"`
	DireccionPublica string `json:"direccion_publica"`
	SistemaOperativo string `json:"sistema_operativo"`
	Dominio          string `json:"dominio"`
}

//Auditoria description
type Auditoria struct {
	ID                 int    `json:"id"`
	Motivo             string `json:"motivo"`
	Comentario         string `json:"comentario"`
	Comandos           string `json:"comandos"`
	Fecha              string `json:"fecha"`
	Usuario            string `json:"usuario"`
	NombreProyecto     string `json:"nombre_proyecto"`
	IPServidor         string `json:"ip_servidor"`
	IDUsuario          int    `json:"id_usuario"`
	IDProyectoServidor int    `json:"id_proyecto_servidor"`
}

//Modificar todos los métodos, querys y structs para que coincidan con lo que se necesita en el front end
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
	usuarios := []Usuario{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if email := c.Query("email"); email != "" {
		selectSQL, _, _ = dialect.From("usuarios").Where(goqu.Ex{
			"email": email,
		}).ToSQL()
	} else {
		selectSQL, _, _ = dialect.From("usuarios").ToSQL()
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, usuarios)
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
		goqu.Record{"email": usuario.Email},
		goqu.Record{"email": usuario.Estado},
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
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rows affected: 0"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	}
	usuario.ID = int(id)
	c.JSON(http.StatusOK, usuario)
}

func getProyectos(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var proyecto Proyecto
	var selectSQL string
	proyectos := []Proyecto{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if nombre := c.Query("nombre"); nombre != "" {
		selectSQL, _, _ = dialect.From("proyectos").Where(goqu.Ex{
			"nombre": nombre,
		}).ToSQL()
	} else {
		selectSQL, _, _ = dialect.From("proyectos").ToSQL()
	}
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
	result, err := db.Exec(updateSQL)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rows affected: 0"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	}
	proyecto.ID = int(id)
	c.JSON(http.StatusOK, proyecto)
}

func getServidores(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var servidor Servidor
	var selectSQL string
	servidores := []Servidor{}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	if direccionPublica := c.Query("direccion_publica"); direccionPublica != "" {
		selectSQL, _, _ = dialect.From("servidores").Where(goqu.Ex{
			"direccion_publica": direccionPublica,
		}).ToSQL()
	} else if dominio := c.Query("dominio"); dominio != "" {
		selectSQL, _, _ = dialect.From("servidores").Where(goqu.Ex{
			"dominio": dominio,
		}).ToSQL()
	} else {
		selectSQL, _, _ = dialect.From("servidores").ToSQL()
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer filas.Close()
	for filas.Next() {
		err = filas.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo, &servidor.Dominio)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		servidores = append(servidores, servidor)
	}
	c.JSON(http.StatusOK, servidores)
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
		err = filas.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo, &servidor.Dominio)
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
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ := dialect.From("servidores").Where(goqu.Ex{
		"id": c.Param("id"),
	}).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo, &servidor.Dominio)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, servidor)
}

func createServidor(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var servidor Servidor
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
	insertSQL, _, _ := dialect.Insert("servidores").Rows(
		goqu.Record{"direccion_publica": servidor.DireccionPublica, "sistema_operativo": servidor.SistemaOperativo, "dominio": servidor.Dominio},
	).ToSQL()
	result, err := db.Exec(insertSQL)
	if err != nil {
		selectSQL, _, _ := dialect.From("servidores").Select("id", "sistema_operativo", "dominio").Where(goqu.Ex{
			"direccion_publica": servidor.DireccionPublica,
		}).ToSQL()
		fila := db.QueryRow(selectSQL)
		err = fila.Scan(&servidor.ID, &servidor.SistemaOperativo, &servidor.Dominio)
		if err != nil {
			selectSQL, _, _ := dialect.From("servidores").Select("id", "direccion_publica", "sistema_operativo").Where(goqu.Ex{
				"dominio": servidor.Dominio,
			}).ToSQL()
			fila := db.QueryRow(selectSQL)
			err = fila.Scan(&servidor.ID, &servidor.DireccionPublica, &servidor.SistemaOperativo)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.Header("Location", "/servidores/"+strconv.Itoa(servidor.ID))
				c.JSON(http.StatusConflict, servidor)
			}
		} else {
			c.Header("Location", "/servidores/"+strconv.Itoa(servidor.ID))
			c.JSON(http.StatusConflict, servidor)
		}
		return
	}
	id, _ := result.LastInsertId()
	servidor.ID = int(id)
	c.JSON(http.StatusCreated, servidor)
}

func updateServidor(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var servidor Servidor
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
		goqu.Record{"direccion_publica": servidor.DireccionPublica, "sistema_operativo": servidor.SistemaOperativo, "dominio": servidor.Dominio},
	).Where(
		goqu.Ex{"id": c.Param("id")},
	).ToSQL()
	result, err := db.Exec(updateSQL)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rows affected: 0"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	}
	servidor.ID = int(id)
	c.JSON(http.StatusOK, servidor)
}

func getAuditorias(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var auditoria Auditoria
	var selectSQL string
	auditorias := []Auditoria{}
	t := time.Now()
	fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
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
			goqu.Ex{"motivo": motivo},
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
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
		).Where(goqu.Ex{
			"email": usuario},
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
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
		).Where(goqu.Ex{
			"nombre": nombreProyecto},
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
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
		).Where(goqu.Ex{
			"direccion_publica": servidorIP},
			goqu.C("fecha").Between(goqu.Range(strings.Replace(fechaLimitInf, "T", " ", -1), strings.Replace(fechaLimitSup, "T", " ", -1))),
		).ToSQL()
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
		).ToSQL()
	}
	filas, err := db.Query(selectSQL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, auditorias)
}

func getAuditoria(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var auditoria Auditoria
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()
	selectSQL, _, _ := dialect.From("auditorias").Where(goqu.Ex{
		"id": c.Param("id"),
	}).ToSQL()
	fila := db.QueryRow(selectSQL)
	err = fila.Scan(&auditoria.ID, &auditoria.Motivo, &auditoria.Comentario, &auditoria.Comandos, &auditoria.Fecha, &auditoria.IDUsuario, &auditoria.IDProyectoServidor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, auditoria)
}

func createAuditoria(c *gin.Context) {
	db, err := obtenerBaseDeDatos()
	var auditoria Auditoria
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
	insertSQL, _, _ := dialect.Insert("auditorias").Rows(
		goqu.Record{"motivo": auditoria.Motivo, "comentario": auditoria.Comentario, "comandos": auditoria.Comandos, "id_usuario": auditoria.IDUsuario, "id_proyecto_servidor": auditoria.IDProyectoServidor},
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
	r.GET("/usuarios", getUsuarios)
	r.GET("/usuarios/:id", getUsuario)
	r.POST("/usuarios", createUsuario)
	r.PUT("/usuarios/:id", updateUsuario)
	r.GET("/proyectos", getProyectos)
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
	r.Run()
}
