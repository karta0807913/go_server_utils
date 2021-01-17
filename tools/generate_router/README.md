# A simple CRUD generator

Install command

```
go get https://github.com/karta0807913/go_server_utils/tools/generate_router
```

## Module requirement

* gin
* gorm
* gopls

## Generate Files

* this package will generate two files, one in the current folder named `./<typename>_<method>.go`,
  other is `../doc/<typename>.go`.


## How to use it

target structure

```golang
type Borrower struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Name string `json:"name" gorm:"not null;index"`
	Phone string `json:"phone" gorm:"not null;index"`
}
```

The generated file will be named `<typename>_<method>.go`

### Create

```
generate_router -type "Borrower" -method "Create"
```

* borrower_create.go

```golang
package model

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// this file generate by go generate, please don't edit it
// data will put into struct
func (insert *Borrower) Create(c *gin.Context, db *gorm.DB) error {
	type Body struct {
		Name  string `json:"name" binding:"required"`
		Phone string `json:"phone" binding:"required"`
	}
	var body Body
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return err
	}

	selectField := []string{
		"name",
		"phone",
	}

	insert.Name = body.Name
	insert.Phone = body.Phone

	return db.Select(
		selectField[0], selectField[1:],
	).Create(&insert).Error
}
```

### Read (find one)

```
generate_router -type "Borrower" -method "First"
```

* borrower_first.go

```golang
package model

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// this file generate by go generate, please don't edit it
// search options will put into struct
func (item *Borrower) First(c *gin.Context, db *gorm.DB) error {
	type Body struct {
		ID    *uint   `form:"id"`
		Name  *string `form:"name"`
		Phone *string `form:"phone"`
	}

	var body Body
	err := c.ShouldBindQuery(&body)
	if err != nil {
		return err
	}

	whereField := make([]string, 0)
	valueField := make([]interface{}, 0)

	if body.ID != nil {
		whereField = append(whereField, "borrowers.id=?")
		valueField = append(valueField, body.ID)
		item.ID = *body.ID
	}

	if body.Name != nil {
		whereField = append(whereField, "borrowers.name=?")
		valueField = append(valueField, body.Name)
		item.Name = *body.Name
	}

	if body.Phone != nil {
		whereField = append(whereField, "borrowers.phone=?")
		valueField = append(valueField, body.Phone)
		item.Phone = *body.Phone
	}

	if len(valueField) == 0 {
		return errors.New("require at least one option")
	}

	err = db.Where(
		strings.Join(whereField, " and "),
		valueField[0], valueField[1:],
	).First(item).Error
	return err
}
```

### Read (find all)

```
generate_router -type "Borrower" -method "Find" -ignore "ID"
```

* borrow_find.go

```golang
package model

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// this file generate by go generate, please don't edit it
// search options will put into struct
func (item *Borrower) Find(c *gin.Context, db *gorm.DB) ([]Borrower, error) {
	type Body struct {
		Name  *string `form:"name"`
		Phone *string `form:"phone"`
	}
	var body Body
	var err error
	_ = c.ShouldBindQuery(&body)

	whereField := make([]string, 0)
	valueField := make([]interface{}, 0)

	if body.Name != nil {
		whereField = append(whereField, "borrowers.name=?")
		valueField = append(valueField, body.Name)
		item.Name = *body.Name
	}

	if body.Phone != nil {
		whereField = append(whereField, "borrowers.phone=?")
		valueField = append(valueField, body.Phone)
		item.Phone = *body.Phone
	}

	var limit int = 20
	slimit, ok := c.GetQuery("limit")
	if ok {
		limit, err = strconv.Atoi(slimit)
		if err != nil {
			limit = 20
		} else {
			if limit <= 0 || 20 < limit {
				limit = 20
			}
		}
	}
	soffset, ok := c.GetQuery("offset")
	var offset int
	if ok {
		offset, err = strconv.Atoi(soffset)
		if err != nil {
			offset = 0
		} else if offset < 0 {
			offset = 0
		}
	} else {
		offset = 0
	}
	var result []Borrower
	if len(whereField) != 0 {
		db = db.Where(
			strings.Join(whereField, " and "),
			valueField[0], valueField[1:],
		)
	}
	err = db.Limit(limit).Offset(offset).Find(&result).Error
	return result, err
}
```

### Update

```
generate_router -type "Borrower" -method "Update"
```

* borrow_update.go

```golang
package model

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// this file generate by go generate, please don't edit it
// data will put into struct
func (insert *Borrower) Update(c *gin.Context, db *gorm.DB) error {
	type Body struct {
		ID uint `json:"id" binding:"required"`

		Name  *string `json:"name"`
		Phone *string `json:"phone"`
	}
	var body Body
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return err
	}
	insert.ID = body.ID

	selectField := make([]string, 0)

	if body.Name != nil {
		selectField = append(selectField, "name")
		insert.Name = *body.Name
	}

	if body.Phone != nil {
		selectField = append(selectField, "phone")
		insert.Phone = *body.Phone
	}

	if len(selectField) == 0 {
		return errors.New("require at least one option")
	}

	return db.Select(
		selectField[0], selectField[1:],
	).Where("borrowers.id=?", body.ID).Updates(&insert).Error
}
```

### Delete

* TODO

## Other parameters

* max_limit
* min_limit

these parameters only used on find method.

it can specify how many results to list

* minItem

Minimal options required
