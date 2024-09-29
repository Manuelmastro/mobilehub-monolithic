package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListAddress(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to get claims",
		})
		return
	}
	CustomClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := CustomClaims.ID
	fmt.Println("id from claims:", userID)

	var Address []responsemodels.Address
	qry := `SELECT * FROM addresses WHERE user_id = ? AND deleted_at IS NULL`
	config.DB.Raw(qry, userID).Scan(&Address)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "address fetched successfully",
		"data": gin.H{
			"address": Address,
		},
	})
}

func AddAddress(c *gin.Context) {

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to get claims",
		})
		return
	}
	CustomClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := CustomClaims.ID
	fmt.Println("id from claims:", userID)

	var Address models.AddressAdd
	err := c.BindJSON(&Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}

	if err := helper.Validate(Address); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	address := models.Addresse{
		UserID:     userID,
		Country:    Address.Country,
		State:      Address.State,
		District:   Address.State,
		StreetName: Address.StreetName,
		PinCode:    Address.PinCode,
		Phone:      Address.Phone,
		Default:    Address.Default,
	}
	config.DB.Create(&address)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "address successfully added",
	})

}

func EditAddress(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to get claims",
		})
		return
	}
	CustomClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := CustomClaims.ID
	fmt.Println("id from claims:", userID)
	AddressID := c.Param("address_id")
	fmt.Println("addressID: ", AddressID)
	var addresscount int
	config.DB.Raw(`SELECT COUNT(*) FROM addresses WHERE id = ? AND user_id = ? and deleted_at IS NULL`, AddressID, userID).Scan(&addresscount)
	if addresscount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "provided address id does not exist"})
		return
	}

	var Address models.AddressAdd
	err := c.BindJSON(&Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}
	if err := helper.Validate(Address); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	address := models.Addresse{
		UserID:     userID,
		Country:    Address.Country,
		State:      Address.State,
		District:   Address.State,
		StreetName: Address.StreetName,
		PinCode:    Address.PinCode,
		Phone:      Address.Phone,
		Default:    Address.Default,
	}

	config.DB.Model(&models.Addresse{}).Where("id = ?", AddressID).Updates(&address)
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "address updated successfully"})

}

func DeleteAddress(c *gin.Context) {

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to get claims",
		})
		return
	}
	CustomClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := CustomClaims.ID
	fmt.Println("id from claims:", userID)
	AddressID := c.Param("address_id")
	fmt.Println("addressID: ", AddressID)
	var addresscount int
	config.DB.Raw(`SELECT COUNT(*) FROM addresses WHERE id = ? AND user_id = ? and deleted_at IS NULL`, AddressID, userID).Scan(&addresscount)
	if addresscount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "provided address id does not exist"})
		return
	}

	config.DB.Where("id = ?", AddressID).Delete(&models.Addresse{})
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "address deleted successfully"})

}
