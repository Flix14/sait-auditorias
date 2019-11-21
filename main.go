package main

import (
	"fmt"

	"github.com/flix14/audit/connection"
	"github.com/flix14/audit/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/doug-martin/goqu/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := connection.ObtenerBaseDeDatos()
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
	r.GET("/usuarios", model.GetUsuarios)
	r.GET("/usuarios/:id", model.GetUsuario)
	r.POST("/usuarios", model.CreateUsuario)
	r.PUT("/usuarios/:id", model.UpdateUsuario)
	r.GET("/proyectos", model.GetProyectos)
	r.GET("/servidores/:id/proyectos", model.GetProyectosPorServidores)
	r.GET("/proyectos/:id", model.GetProyecto)
	r.POST("/proyectos", model.CreateProyecto)
	r.PUT("/proyectos/:id", model.UpdateProyecto)
	r.GET("/servidores", model.GetServidores)
	r.GET("/proyectos/:id/servidores", model.GetServidoresPorProyecto)
	r.GET("/servidores/:id", model.GetServidor)
	r.POST("/servidores", model.CreateServidor)
	r.PUT("/servidores/:id", model.UpdateServidor)
	r.GET("/auditorias", model.GetAuditorias)
	r.GET("/auditorias/:id", model.GetAuditoria)
	r.POST("/auditorias", model.CreateAuditoria)
	r.Run(":3000")
}
