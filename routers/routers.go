package routers

import (
	"strconv"

	"github.com/K31NER/url-shortener/models"
	"github.com/K31NER/url-shortener/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Manejador de rutas
func SetupRoutes(app *fiber.App, conn *gorm.DB) {
	// Aquí definiremos nuestras rutas 

    app.Get("/api/v1/:short_url",func (c *fiber.Ctx) error {
		return short_urlHandler(c, conn)
	})

	app.Get("/api/v1/urls/list",func (c *fiber.Ctx) error {
		return readUrls(c, conn)
	})

    app.Post("/api/v1", func (c *fiber.Ctx) error {
		return addUrlHandler(c,conn)
	})

	app.Delete("/api/v1/:id", func (c *fiber.Ctx) error {
        return deleteUrls(c,conn)
	})
}

// Definimos el validador
var validate = validator.New()

// ---- Funciones de cada ruta ---- // 

// Funcion de redireccion
func short_urlHandler(c *fiber.Ctx, db *gorm.DB) error {

	short_url := c.Params("short_url")
    
	// Buscamos y aumentamos los clicks
	original_url,err := utils.ManageVisit(short_url,db)
    
	// Validamos si no fallo nada
	if err != nil{
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":"Url no encontrada",
		})
	} 
    
	return c.Redirect(original_url,fiber.StatusFound)
}

// Crear nuevo link recortado
func addUrlHandler(c *fiber.Ctx, db *gorm.DB) error{
    
	// Definimos el body que esperamos
	var body models.JsonURLInfo
	
	// Validamos si hubo un error de parseo
	if err := c.BodyParser(&body); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"Json invalido",
		})
	}
    
	// Validamos los campos y su estructura
	if err := validate.Struct(body); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Creamos la url sin el short url
	url := models.URLTable{
		OriginalURL: body.OriginalURL,
	}
    
	// Creamos a la vez que validamos cualquier error
	if err := db.Create(&url).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo guardar la URL",
		})
	}
    
	// Generamos el id para la url recortada
	short := utils.CreatShortID(uint64(url.ID))
    
	// Actualizamos el registro
	if err := db.Model(&url).Update("short_url",short).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo generar short_url",
		})
	}

	// Refrescamose short
	url.ShortURL = short
    
	// Devolvemos el objeto
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":url,
		"shortener_url": "http://127.0.0.1:8080/api/v1/"+url.ShortURL,
	})
	
}

// Listar todos los links en la base de datos
func readUrls(c *fiber.Ctx, db *gorm.DB) error{
    
	// Obtenemos los registros de la base de datos
	urls, err := utils.ReadAllUrls(db)

	// Verificamos que no tenga error
	if err != nil{
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al obtener urls: " + err.Error(),
			})
	}

	return c.Status(fiber.StatusOK).JSON(urls)
}

func deleteUrls(c *fiber.Ctx, db *gorm.DB) error{
    
	// Convertimos a entero
	id, err  := strconv.ParseInt(c.Params("id"),10,64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El id de ser solo enteros",
		})
	}
	
	if err := utils.DeleteUrl(id,db) ;err != nil {
		return  c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No se logro eliminar la url",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":"url eliminada con exito",
	})
}