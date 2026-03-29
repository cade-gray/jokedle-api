package routes

import (
	"fmt"
	"jokedle-api/middleware"
	"jokedle-api/models"
	"math/rand"
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

		// GET /joke/sequence - Get sequence information (requires authentication)
		jokeProtected.GET("/sequence", getSequence(db))

		// POST /joke/sequence - Update sequence number (requires authentication)
		jokeProtected.POST("/sequence", updateSequence(db))

		// GET /joke/submission/all - Get all joke submissions (requires authentication)
		jokeProtected.GET("/submission/all", getAllJokeSubmissions(db))
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
			fmt.Println(result.Error)
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
			Joke struct {
				Setup     string  `json:"setup"`
				Punchline string  `json:"punchline"`
				Source    *string `json:"source"`
			} `json:"joke"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// Create JokeSubmission without setting SubmissionID
		joke := models.JokeSubmission{
			Setup:     requestBody.Joke.Setup,
			Punchline: requestBody.Joke.Punchline,
			Source:    requestBody.Joke.Source,
		}

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

		// Create the submission - PostgreSQL will auto-generate the ID
		result := db.Create(&joke)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to submit joke",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"message":      "Joke submitted successfully",
			"submissionId": joke.SubmissionID, // Return the generated ID
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
			SequenceNbr *int `json:"sequenceNbr"` // Changed to pointer to detect nil/absence
			RandomTf    bool `json:"randomTf"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}
		// Reject if both sequenceNbr and randomTf are provided
		if requestBody.SequenceNbr != nil && requestBody.RandomTf {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Cannot specify both sequenceNbr and randomTf. Choose either a specific sequence number or random generation.",
			})
			return
		}

		// Pull the total count of jokes to validate ranges
		var count int64
		countLookupResult := db.Model(&models.Joke{}).Count(&count)
		if countLookupResult.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count jokes"})
			return
		}

		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "No jokes available to sequence",
			})
			return
		}

		var sequenceNumber int

		// Determine which sequence number to use
		if requestBody.SequenceNbr != nil {
			// Use provided sequence number
			sequenceNumber = *requestBody.SequenceNbr

			// Validate the provided sequence number
			if int64(sequenceNumber) > count || sequenceNumber < 1 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Invalid sequence number. Must be between 1 and %d", count),
				})
				return
			}
		} else if requestBody.RandomTf {
			// Generate random sequence number
			sequenceNumber = 1 + int(rand.Int63n(count))
		} else {
			// Neither sequenceNbr provided nor random requested
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Either sequenceNbr must be provided or randomTf must be true",
			})
			return
		}

		// Reject if both sequenceNbr and randomTf are provided
		if requestBody.SequenceNbr != nil && requestBody.RandomTf {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Cannot specify both sequenceNbr and randomTf. Choose either a specific sequence number or random generation.",
			})
			return
		}

		// Update the sequence
		result := db.Model(&models.Sequence{}).
			Where("sequence_name = ?", "JokeOfDay").
			Update("sequence_nbr", sequenceNumber)

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

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"sequenceNbr": sequenceNumber,
			"randomTf":    requestBody.RandomTf,
		})
	}
}
