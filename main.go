package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Series representa una serie con sus atributos.
type Series struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	CurrentEpisode int    `json:"current_episode"`
	TotalEpisodes  int    `json:"total_episodes"`
	Status         string `json:"status"`
	Score          int    `json:"score"`
}

var db *sql.DB

// initDB inicializa la conexión a la base de datos usando variables de entorno.
func initDB() {
	var err error
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error conectándose a la base de datos: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Error al hacer ping a la base de datos: %v", err)
	}
}

func main() {
	initDB()

	router := gin.Default()

	// Registro de endpoints de la API
	router.GET("/api/series", getAllSeries)
	router.GET("/api/series/:id", getSeriesByID)
	router.POST("/api/series", createSeries)
	router.PUT("/api/series/:id", updateSeries)
	router.DELETE("/api/series/:id", deleteSeries)
	router.PATCH("/api/series/:id/status", updateSeriesStatus)
	router.PATCH("/api/series/:id/episode", incrementEpisode)
	router.PATCH("/api/series/:id/upvote", upvoteSeries)
	router.PATCH("/api/series/:id/downvote", downvoteSeries)

	// Levanta el servidor en el puerto 8080
	router.Run(":8080")
}

// getAllSeries obtiene todas las series de la base de datos.
func getAllSeries(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, description, current_episode, total_episodes, status, score FROM series")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var seriesList []Series
	for rows.Next() {
		var s Series
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.CurrentEpisode, &s.TotalEpisodes, &s.Status, &s.Score); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		seriesList = append(seriesList, s)
	}
	c.JSON(http.StatusOK, seriesList)
}

// getSeriesByID obtiene una serie por su ID.
func getSeriesByID(c *gin.Context) {
	id := c.Param("id")
	var s Series
	err := db.QueryRow("SELECT id, title, description, current_episode, total_episodes, status, score FROM series WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Description, &s.CurrentEpisode, &s.TotalEpisodes, &s.Status, &s.Score)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, s)
}

// createSeries crea una nueva serie.
func createSeries(c *gin.Context) {
	var newSeries Series
	if err := c.BindJSON(&newSeries); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `INSERT INTO series (title, description, current_episode, total_episodes, status, score)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := db.QueryRow(query, newSeries.Title, newSeries.Description, newSeries.CurrentEpisode, newSeries.TotalEpisodes, newSeries.Status, newSeries.Score).
		Scan(&newSeries.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newSeries)
}

// updateSeries actualiza completamente los datos de una serie existente.
func updateSeries(c *gin.Context) {
	id := c.Param("id")
	var s Series
	if err := c.BindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `UPDATE series SET title=$1, description=$2, current_episode=$3, total_episodes=$4, status=$5, score=$6 WHERE id=$7`
	result, err := db.Exec(query, s.Title, s.Description, s.CurrentEpisode, s.TotalEpisodes, s.Status, s.Score, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Serie actualizada exitosamente"})
}

// deleteSeries elimina una serie por ID.
func deleteSeries(c *gin.Context) {
	id := c.Param("id")
	result, err := db.Exec("DELETE FROM series WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Serie eliminada exitosamente"})
}

// updateSeriesStatus actualiza el estado de una serie.
func updateSeriesStatus(c *gin.Context) {
	id := c.Param("id")
	var payload struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.Exec("UPDATE series SET status = $1 WHERE id = $2", payload.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Estado actualizado exitosamente"})
}

// incrementEpisode incrementa en 1 el contador de episodios actuales.
func incrementEpisode(c *gin.Context) {
	id := c.Param("id")
	result, err := db.Exec("UPDATE series SET current_episode = current_episode + 1 WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		return
	}
	// Se devuelve la serie actualizada.
	var s Series
	err = db.QueryRow("SELECT id, title, description, current_episode, total_episodes, status, score FROM series WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Description, &s.CurrentEpisode, &s.TotalEpisodes, &s.Status, &s.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// upvoteSeries incrementa la puntuación de una serie en 1.
func upvoteSeries(c *gin.Context) {
	id := c.Param("id")
	result, err := db.Exec("UPDATE series SET score = score + 1 WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		return
	}
	// Se devuelve la serie actualizada.
	var s Series
	err = db.QueryRow("SELECT id, title, description, current_episode, total_episodes, status, score FROM series WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Description, &s.CurrentEpisode, &s.TotalEpisodes, &s.Status, &s.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// downvoteSeries disminuye la puntuación de una serie en 1.
func downvoteSeries(c *gin.Context) {
	id := c.Param("id")
	result, err := db.Exec("UPDATE series SET score = score - 1 WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Serie no encontrada"})
		return
	}
	// Se devuelve la serie actualizada.
	var s Series
	err = db.QueryRow("SELECT id, title, description, current_episode, total_episodes, status, score FROM series WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Description, &s.CurrentEpisode, &s.TotalEpisodes, &s.Status, &s.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}