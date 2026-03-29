package models

import (
	"time"
)

// Joke represents the main jokes table
type Joke struct {
	JokeID             int       `json:"jokeId" gorm:"primaryKey;column:jokeid;autoIncrement"`
	Setup              string    `json:"setup" gorm:"type:varchar(255);not null"`
	Punchline          string    `json:"punchline" gorm:"type:varchar(50);not null"`
	FormattedPunchline string    `json:"formattedPunchline" gorm:"type:text;column:formattedpunchline;not null"`
	Source             *string   `json:"source" gorm:"type:varchar(45)"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// Sequence represents the sequences table for managing joke sequences
type Sequence struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	SequenceName string    `json:"sequenceName" gorm:"type:varchar(100);not null;uniqueIndex"`
	SequenceNbr  int       `json:"sequenceNbr" gorm:"not null"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// JokeSubmission represents jokes submitted by users for approval
type JokeSubmission struct {
	SubmissionID int       `json:"submissionId" gorm:"primaryKey;column:submissionid;autoIncrement"`
	Setup        string    `json:"setup" gorm:"type:varchar(255);not null"`
	Punchline    string    `json:"punchline" gorm:"type:varchar(50);not null"`
	Source       *string   `json:"source" gorm:"type:varchar(45)"`
	CreatedAt    time.Time `json:"createdAt"`
}

// JokeCount represents the response for joke count queries
type JokeCount struct {
	Count int64 `json:"count"`
}

// JokeWebList represents the simplified joke list for web display
type JokeWebList struct {
	JokeID int    `json:"jokeId" gorm:"primaryKey;column:jokeid"`
	Setup  string `json:"setup"`
}

// Table names for GORM
func (Joke) TableName() string {
	return "jokes"
}

func (Sequence) TableName() string {
	return "sequences"
}

func (JokeSubmission) TableName() string {
	return "jokesubmission"
}
