package model

type Category struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(50);not null" json:"name"`
	ParentID *int64 `gorm:"index" json:"parent_id"`
	IconURL  string `gorm:"type:varchar(255)" json:"icon_url"`

	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (c *Category) TableName() string { return "categories" }

type StoreCategory struct {
	StoreID    int64 `gorm:"primaryKey" json:"store_id"`
	CategoryID int64 `gorm:"primaryKey" json:"category_id"`
}

func (sc *StoreCategory) TableName() string { return "store_categories" }
