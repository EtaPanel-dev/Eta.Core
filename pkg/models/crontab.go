package models

type CrontabEntry struct {
	Id       int    `json:"id"`
	Minute   string `json:"minute"`
	Hour     string `json:"hour"`
	Day      string `json:"day"`
	Month    string `json:"month"`
	Weekday  string `json:"weekday"`
	Command  string `json:"command"`
	Comment  string `json:"comment"`
	Enabled  bool   `json:"enabled"`
	FullLine string `json:"fullLine"`
	NextRun  string `json:"nextRun"`
}

type CrontabRequest struct {
	Minute  string `json:"minute" binding:"required"`
	Hour    string `json:"hour" binding:"required"`
	Day     string `json:"day" binding:"required"`
	Month   string `json:"month" binding:"required"`
	Weekday string `json:"weekday" binding:"required"`
	Command string `json:"command" binding:"required"`
	Comment string `json:"comment"`
	Enabled bool   `json:"enabled"`
}
