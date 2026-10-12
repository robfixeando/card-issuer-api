package handlers

import (
	"card-issuer-api/database"
	"card-issuer-api/models"
	"card-issuer-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateCardInput struct {
	CardholderName string `json:"cardholder_name" binding:"required"`
	Pan            string `json:"pan" binding:"required,min=13,max=19"`
}

type BatchCardInput struct {
	Cards []CreateCardInput `json:"cards" binding:"required,dive"`
}

func CreateCard(c *gin.Context) {
	// Obtener el tenant_id inyectado por el middleware
	tenantID := c.GetString("tenant_id")

	var input CreateCardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extraer últimos 4 dígitos y hashear el PAN completo
	last4 := input.Pan[len(input.Pan)-4:]
	panHash := utils.HashPan(input.Pan)

	card := models.Card{
		TenantID:       tenantID,
		CardholderName: input.CardholderName,
		Last4:          last4,
		PanHash:        panHash,
		Status:         "ISSUED",
	}

	if err := database.DB.Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo emitir la tarjeta"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tarjeta emitida exitosamente",
		"data":    card,
	})
}

func BatchCreateCards(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	var input BatchCardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var createdCards []models.Card

	// Transacción para asegurar la consistencia
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range input.Cards {
			last4 := item.Pan[len(item.Pan)-4:]
			panHash := utils.HashPan(item.Pan)

			card := models.Card{
				TenantID:       tenantID,
				CardholderName: item.CardholderName,
				Last4:          last4,
				PanHash:        panHash,
				Status:         "ISSUED",
			}

			if err := tx.Create(&card).Error; err != nil {
				return err
			}
			createdCards = append(createdCards, card)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error procesando el lote de tarjetas"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Lote procesado exitosamente",
		"total":   len(createdCards),
		"data":    createdCards,
	})
}

func ListCards(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	statusQuery := c.Query("status")

	var cards []models.Card
	var total int64

	query := database.DB.Model(&models.Card{}).Where("tenant_id = ?", tenantID)

	if statusQuery != "" {
		query = query.Where("status = ?", statusQuery)
	}

	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron recuperar las tarjetas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  cards,
	})
}

func UpdateCardStatus(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	cardID := c.Param("id")

	var input struct {
		Status string `json:"status" binding:"required,oneof=ISSUED BLOCKED REVOKED"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card models.Card
	// Busco la tarjeta asegurando que pertenezca al tenant actual
	if err := database.DB.Where("id = ? AND tenant_id = ?", cardID, tenantID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarjeta no encontrada"})
		return
	}

	card.Status = input.Status
	if err := database.DB.Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el estado de la tarjeta"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estado de la tarjeta actualizado exitosamente",
		"data":    card,
	})
}

func GetCardByID(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	cardID := c.Param("id")

	var card models.Card

	if err := database.DB.Where("id = ? AND tenant_id = ?", cardID, tenantID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarjeta no encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": card,
	})
}
