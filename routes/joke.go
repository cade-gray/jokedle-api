package routes

import (
	"jokedle-api/middleware"
	"jokedle-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterJokeRoutes registers all joke-related routes with their handlers
func RegisterJokeRoutes(router *gin.Engine, db *gorm.DB, authDB *gorm.DB) {
	jokeGroup := router.Group("/joke")
	{
		// GET /joke - Get jokes that are in sequences (public endpoint)
		jokeGroup.GET("", getJokesInSequence(db))

		// GET /joke/id/:id - Get a specific joke by ID (public endpoint)
		jokeGroup.GET("/id/:id", getJokeByID(db))

		// GET /joke/count - Get total count of jokes (public endpoint)
		jokeGroup.GET("/count", getJokeCount(db))

		// GET /joke/all/weblist - Get simplified joke list for web display (public endpoint)
		jokeGroup.GET("/all/weblist", getJokeWebList(db))

		// POST /joke/submission - Submit a joke for approval (public endpoint)
		jokeGroup.POST("/submission", submitJoke(db))
	}

	// Protected routes that require authentication
	jokeProtected := router.Group("/joke")
	jokeProtected.Use(middleware.AuthenticateToken(authDB))
	{
		// GET /joke/all - Get all jokes (requires authentication)
		jokeProtected.GET("/all", getAllJokes(db))

		// POST /joke - Create a new joke (requires authentication)
		jokeProtected.POST("", createJoke(db))

		// POST /joke/getsequence - Get sequence information (requires authentication)
		jokeProtected.POST("/getsequence", getSequence(db))

		// POST /joke/updatesequence - Update sequence number (requires authentication)
		jokeProtected.POST("/updatesequence", updateSequence(db))

		// POST /joke/submission/all - Get all joke submissions (requires authentication)
		jokeProtected.POST("/submission/all", getAllJokeSubmissions(db))
	}
}

// getJokesInSequence handles GET /joke - Returns jokes that are in the sequence
func getJokesInSequence(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var jokes []models.Joke

		// Query jokes that have matching IDs in the sequences table
		result := db.Where("jokeid IN (SELECT sequence_nbr FROM sequences)").Find(&jokes)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve jokes from sequence"})
			return
		}

		c.JSON(http.StatusOK, jokes)
	}
}

// getJokeByID handles GET /joke/id/:id - Returns a specific joke by ID
func getJokeByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid joke ID"})
			return
		}

		var joke models.Joke
		result := db.Where("jokeid = ?", id).First(&joke)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Joke not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve joke"})
			return
		}

		c.JSON(http.StatusOK, joke)
	}
}

// getAllJokes handles POST /joke/all - Returns all jokes (requires authentication)
func getAllJokes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var jokes []models.Joke

		result := db.Order("jokeid DESC").Find(&jokes)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve jokes"})
			return
		}

		c.JSON(http.StatusOK, jokes)
	}
}

// getJokeWebList handles GET /joke/all/weblist - Returns simplified joke list for web display
func getJokeWebList(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var jokes []models.JokeWebList

		result := db.Model(&models.Joke{}).Order("jokeId DESC").Select("jokeId, setup").Find(&jokes)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve joke list"})
			return
		}
		c.JSON(http.StatusOK, jokes)
	}
}

// createJoke handles POST /joke - Creates a new joke (requires authentication)
func createJoke(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			Joke models.Joke `json:"joke"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		joke := requestBody.Joke

		// Validate required fields
		if joke.Setup == "" || joke.Punchline == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Setup and punchline are required",
			})
			return
		}

		// Validate field lengths
		if len(joke.Setup) > 255 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   "Setup exceeded character limit (255). Please adjust accordingly.",
			})
			return
		}
		if len(joke.Punchline) > 50 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   "Punchline exceeded character limit (50). Please adjust accordingly.",
			})
			return
		}

		// Get the next joke ID
		var maxID int
		db.Model(&models.Joke{}).Select("COALESCE(MAX(jokeid), 0)").Scan(&maxID)
		joke.JokeID = maxID + 1

		// Create the joke
		result := db.Create(&joke)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to create joke",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Joke inserted successfully",
		})
	}
}

// getSequence handles POST /joke/getsequence - Returns sequence information (requires authentication)
func getSequence(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sequences []models.Sequence

		result := db.Where("sequence_name = ?", "JokeOfDay").Find(&sequences)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sequences"})
			return
		}

		c.JSON(http.StatusOK, sequences)
	}
}

// submitJoke handles POST /joke/submission - Submits a joke for approval
func submitJoke(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			Joke models.JokeSubmission `json:"joke"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		joke := requestBody.Joke

		// Validate field lengths
		if len(joke.Setup) > 255 || len(joke.Punchline) > 50 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   "Setup, Punchline, or Source exceeded character limit. Please adjust accordingly.",
			})
			return
		}

		if joke.Source != nil && len(*joke.Source) > 45 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   "Source exceeded character limit (45). Please adjust accordingly.",
			})
			return
		}

		// Create the submission
		result := db.Create(&joke)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to submit joke",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Joke submitted successfully",
		})
	}
}

// getAllJokeSubmissions handles POST /joke/submission/all - Returns all joke submissions (requires authentication)
func getAllJokeSubmissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var submissions []models.JokeSubmission

		result := db.Order("submissionid DESC").Find(&submissions)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve joke submissions"})
			return
		}

		c.JSON(http.StatusOK, submissions)
	}
}

// updateSequence handles POST /joke/updatesequence - Updates the sequence number (requires authentication)
func updateSequence(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			SequenceNbr int `json:"sequenceNbr"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		result := db.Model(&models.Sequence{}).
			Where("sequence_name = ?", "JokeOfDay").
			Update("sequence_nbr", requestBody.SequenceNbr)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to update sequence",
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "No sequence found to update",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

// getJokeCount handles GET /joke/count - Returns the total count of jokes
func getJokeCount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int64

		result := db.Model(&models.Joke{}).Count(&count)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count jokes"})
			return
		}

		c.JSON(http.StatusOK, []gin.H{{"count": count}})
	}
}
