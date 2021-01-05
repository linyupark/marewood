package models

import (
	"MareWood/helper"
	"MareWood/sql"
	"errors"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
)

var fillWhiteListForCategory = []string{"desc", "name"}

type Category struct {
	gorm.Model
	Name        string `binding:"required,min=2,max=20"`
	JobQuantity int    `gorm:"default:0"`
	Desc        string `gorm:"type:varchar(1000)"`
}

func (c *Category) CategoryFindAll(keyword string) (categories []Category, err error) {
	query := sql.DB.Order("created_at asc")
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+ keyword +"%")
	}
	err = query.Find(&categories).Error
	return
}

func (c *Category) Create() (err error) {
	c.JobQuantity = 0
	err = sql.DB.Create(&c).Error
	return
}

func (c *Category) UpdateFieldContent(id string, field string, fieldContent string) (err error) {

	field = strings.ToLower(field)

	if !helper.InStrArr(field, fillWhiteListForCategory) {
		return errors.New("不能修改当前字段！")
	}

	return sql.DB.Model(&c).Where("id = ?", id).Update(field, fieldContent).Error

}

func (c *Category) CategoryJobQuantityIncrement(id int) (err error) {
	err =
		sql.DB.Model(&c).Where("id = ?", id).
			UpdateColumn("job_quantity", gorm.Expr("job_quantity + ?", 1)).Error
	return
}

func (c *Category) CategoryJobQuantityDecrement(id int) (err error) {
	err =
		sql.DB.Model(&c).Where("id = ?", id).
			UpdateColumn("job_quantity", gorm.Expr("job_quantity - ?", 1)).Error
	return
}

func (c *Category) Destroy(id string) (err error) {

	var jobCount int

	err = sql.DB.Model(&Job{}).
		Where("category_id = ?", id).
		Count(&jobCount).Error

	if err != nil {
		return err
	}
	if jobCount > 0 {
		return errors.New("can't delete，There are " + strconv.Itoa(jobCount) + " more tasks using this category")
	}
	return sql.DB.Where("id = ?", id).Delete(&c).Error
}
