package models

type User struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Username  string  `json:"username" gorm:"not null"`
	Email     string  `json:"email" gorm:"unique;not null"`
	Password  string  `json:"-" gorm:"not null"`
	Role      string  `json:"role" gorm:"type:varchar(20);default:'user'"`
	CompanyID uint    `json:"companyId" gorm:"not null"`
	Company   Company `json:"company" gorm:"foreignKey:CompanyID"`
	Items     []Item  `json:"items" gorm:"foreignKey:UserID"`
	Verified  bool    `json:"verified" gorm:"default:false"`
}
