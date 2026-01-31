package model

type Tag struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	PostCount int    `gorm:"default:0" json:"post_count"`
}

func (t *Tag) TableName() string {
	return "tags"
}
