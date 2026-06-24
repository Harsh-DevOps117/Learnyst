package models

import "gorm.io/gorm"


type User struct {
	gorm.Model
	Id         			int    `gorm:"primaryKey; autoIncrement"`
	First_Name 			string `gorm:"type:varchar(50); not null"`
	Middle_Name 		string `gorm:"type:varchar(50); not null"`
	Last_Name 			string `gorm:"type:varchar(50); not null"`

	Email 					string `gorm:"type:varchar(255); unique; not null uniqueIndex"`
	Phone 					string `gorm:"type:varchar(20);unique; not null uniqueIndex"`
	Email_Verified 	bool 	 `gorm:"default:false"`
	Phone_Verified 	bool 	 `gorm:"default:false"`

	Subscriptions 	[]Subscription `gorm:"foreignKey:Owner_User_id"`
	Roles 					Role   `gorm:"foreignKey:Id"`

	Is_Active 		 	bool 	 `gorm:"default:true"`
	Location 				string `gorm:"type:text; not null"`
	Password 				string `gorm:"type:varchar(255); not null"`
}

type Refresh_Token struct{
	Id 						int	`gorm:"primaryKey; autoIncrement"`
	Token 				string `gorm:"type:text; not null"`
	Owner_User_id User `gorm:"foreignKey:Owner_User_id"`
	Device_Name 	string `gorm:"type:varchar(50); not null"`
	IP_Address  	string `gorm:"type:varchar(50); not null"`
	Created_At  	string `gorm:"type:varchar(50); not null default: CURRENT_TIMESTAMP"`
	Expires_At  	string `gorm:"type:varchar(50); default: 7*24*60*60"`
	Revoked_At  	string `gorm:"type:varchar(50); default: null"`
}

type Permission struct{
	Id 							int	`gorm:"primaryKey; autoIncrement"`
	Role            Role   `gorm:"foreignKey:Id"`
	Permission_Type string `gorm:"type:varchar(50); not null"`
	Permission_Name string `gorm:"type:varchar(50); not null"`
}

type Role struct{
	Id 			int	`gorm:"primaryKey; autoIncrement"`
	Role_Name string `gorm:"type:varchar(50); not null"`
}

type Subscription struct{
	gorm.Model
	Owner_User_id User 	`gorm:"foreignKey:Owner_User_id"`
	Tenent_id    	Tenent `gorm:"foreignKey:Tenent_id"`
	Plan_Name    	string `gorm:"type:varchar(50); not null"`
	Plan_Type    	string `gorm:"type:varchar(50); not null"`
	Plan_Status  	string `gorm:"type:varchar(50); not null"`
	Plan_Started 	string `gorm:"type:varchar(50); not null"`
	Plan_Ended   	string `gorm:"type:varchar(50); not null"`
	Plan_Cost    	string `gorm:"type:varchar(50); not null"`
	Plan_Duration 	string `gorm:"type:varchar(50); not null"`
}

type Tenent struct{
	gorm.Model
	Owner_User_id User 						`gorm:"foreignKey:Owner_User_id"`
	Academy_Name 	string 					`gorm:"type:varchar(100); not null"`
	Academy_Slug 	string 					`gorm:"type:varchar(100); not null"`
	Description  	string 					`gorm:"type:text;"`
	Logo_Url     	string 					`gorm:"type:text;"`
	Banner_Url  	string 					`gorm:"type:text;"`
	Custum_Domain string 					`gorm:"type:varchar(100); not null"`
	Subscription_Plan 	[]Subscription 	`gorm:"foreignKey:Tenent_id"`
	Is_Active 			bool 					`gorm:"default:true"`
	IS_Suspended 		bool 					`gorm:"default:false"`
}
