package utils

import (
	"github.com/K31NER/url-shortener/models"

	"go.osspkg.com/x/algorithms/encoding/base62"
	"gorm.io/gorm"
)

func CreatShortID(id uint64) string{
    
	// Definimos el codificador base 62
	encoder := base62.New("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

	// Codificamos el id
	encodeID := encoder.Encode(id)
    
	// Devolvemos
	return encodeID
}

// Busca la url y aumenta el numero de clicks
func ManageVisit(short_url string, db *gorm.DB) (string, error) {
    var url models.URLTable
    var originalURL string

    // Transaction maneja automáticamente el Begin, Commit y Rollback
    err := db.Transaction(func(tx *gorm.DB) error {

        // 1. Buscar
        if err := tx.Where("short_url = ?", short_url).First(&url).Error; err != nil {
            return err // Esto hace Rollback automático
        }

        // 2. Actualizar
        if err := tx.Model(&url).Update("clicks", gorm.Expr("clicks + ?", 1)).Error; err != nil {
            return err // Esto hace Rollback automático
        }

        originalURL = url.OriginalURL // Pasamos la url de origin
		
        return nil // Esto hace Commit automático
    })

    return originalURL, err
}


// Devuelve todos las urls registradas
func ReadAllUrls(db *gorm.DB) ([]models.URLTable,error)  {
	var urls []models.URLTable

	// Realizamos la busqueda en la base de datos
	result := db.Find(&urls)
    
	// Validamos que no tenga error
	if result.Error != nil {
		return nil, result.Error
	}
	
	return  urls, nil
}

// elimina las urls
func DeleteUrl(id int64, db *gorm.DB) error {
    
	// Buscamos el url y lo borramos
	result := db.Delete(&models.URLTable{}, id)
    
	// Validamos si encontro
	if result.Error != nil{
		return  result.Error
	}
    
	// Validamos si logro borrar
	if result.RowsAffected == 0 {
		return  gorm.ErrRecordNotFound
	}

	return  nil
}