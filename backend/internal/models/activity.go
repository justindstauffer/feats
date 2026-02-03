package models

import (
	"time"
)

type ActivityType struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	Name      string    `gorm:"type:text;uniqueIndex;not null" json:"name"`
	Icon      *string   `gorm:"type:text" json:"icon,omitempty"`
	IsSystem  bool      `gorm:"not null;default:false" json:"is_system"`
	CreatedBy *string   `gorm:"type:text" json:"created_by,omitempty"`
	CreatedAt time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	Creator *User  `gorm:"foreignKey:CreatedBy" json:"-"`
	Posts   []Post `gorm:"foreignKey:ActivityTypeID" json:"-"`
}

func (ActivityType) TableName() string {
	return "activity_types"
}

// CoreActivityTypes returns the default activity types to seed
func CoreActivityTypes() []ActivityType {
	return []ActivityType{
		{Name: "Gym", Icon: stringPtr("🏋️"), IsSystem: true},
		{Name: "Hiking", Icon: stringPtr("🥾"), IsSystem: true},
		{Name: "Golf", Icon: stringPtr("⛳"), IsSystem: true},
		{Name: "Walking", Icon: stringPtr("🚶"), IsSystem: true},
		{Name: "Running", Icon: stringPtr("🏃"), IsSystem: true},
		{Name: "Cycling", Icon: stringPtr("🚴"), IsSystem: true},
		{Name: "Swimming", Icon: stringPtr("🏊"), IsSystem: true},
		{Name: "Achievement", Icon: stringPtr("🏆"), IsSystem: true},
	}
}

func stringPtr(s string) *string {
	return &s
}
