package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"card-issuer-api/database"
	"card-issuer-api/middleware"
	"card-issuer-api/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	database.DB, _ = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	database.DB.AutoMigrate(&models.Card{})

	r := gin.Default()
	api := r.Group("/api/v1")
	api.Use(middleware.TenantMiddleware())
	{
		api.POST("/cards", CreateCard)
		api.GET("/cards", ListCards)
		api.GET("/cards/:id", GetCardByID)
		api.PATCH("/cards/:id/status", UpdateCardStatus)
	}

	return r
}

func TestCreateCard(t *testing.T) {
	r := setupTestRouter()

	body := map[string]string{
		"cardholder_name": "Juan Perez",
		"pan":             "4532111122223333",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/cards", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "banco-galicia")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Esperaba status 201, obtuve %d", w.Code)
	}
}

func TestTenantDataIsolation(t *testing.T) {
	r := setupTestRouter()

	database.DB.Create(&models.Card{
		TenantID:       "banco-galicia",
		CardholderName: "Alice Galicia",
		Last4:          "1111",
		PanHash:        "hash_galicia",
		Status:         "ISSUED",
	})

	database.DB.Create(&models.Card{
		TenantID:       "banco-macro",
		CardholderName: "Bob Macro",
		Last4:          "2222",
		PanHash:        "hash_macro",
		Status:         "ISSUED",
	})

	// 1. Verificar listado aislado
	reqList, _ := http.NewRequest("GET", "/api/v1/cards", nil)
	reqList.Header.Set("X-Tenant-ID", "banco-galicia")
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)

	var response map[string]interface{}
	json.Unmarshal(wList.Body.Bytes(), &response)
	if response["total"].(float64) != 1 {
		t.Errorf("El tenant 'banco-galicia' debería ver exactamente 1 tarjeta")
	}

	// 2. Verificar GET /cards/:id ajeno (Debe retornar 404)
	reqGet, _ := http.NewRequest("GET", "/api/v1/cards/2", nil)
	reqGet.Header.Set("X-Tenant-ID", "banco-galicia")
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusNotFound {
		t.Errorf("Esperaba 404 al consultar tarjeta de otro tenant, obtuve %d", wGet.Code)
	}
}

func TestListCardsPaginationAndFilter(t *testing.T) {
	r := setupTestRouter()

	// Insertar tarjetas de prueba para un tenant
	database.DB.Create(&models.Card{TenantID: "banco-galicia", CardholderName: "User 1", Last4: "1111", PanHash: "h1", Status: "ISSUED"})
	database.DB.Create(&models.Card{TenantID: "banco-galicia", CardholderName: "User 2", Last4: "2222", PanHash: "h2", Status: "BLOCKED"})

	// Probar filtro por status BLOCKED
	req, _ := http.NewRequest("GET", "/api/v1/cards?status=BLOCKED", nil)
	req.Header.Set("X-Tenant-ID", "banco-galicia")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["total"].(float64) != 1 {
		t.Errorf("Esperaba 1 tarjeta con status BLOCKED, obtuve %f", response["total"].(float64))
	}
}
