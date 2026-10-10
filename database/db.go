package database

import (
	"card-issuer-api/models"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error
	fmt.Println("[DB] Abriendo conexión a SQLite (modo puro Go)...")

	DB, err = gorm.Open(sqlite.Open("issuer.db"), &gorm.Config{})
	if err != nil {
		panic("Error al conectar con la base de datos: " + err.Error())
	}

	fmt.Println("[DB] Conexión exitosa. Ejecutando AutoMigrate...")
	DB.AutoMigrate(&models.Card{})

	fmt.Println("[DB] Migración finalizada correctamente.")
}
