package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 定义gorm model
type Product struct {
	ID    uint   `gorm:"primarykey"`
	Code  string `gorm:"column:code"`
	Price uint   `gorm:"column:user_id"`
}

// 为model定义表名
func (p Product) TableName() string {
	return "product"
}

type User struct {
	ID      int64
	Name    string `gorm:"default:galeone"`
	Age     int64  `gorm:"default:18"`
	Deleted gorm.DeletedAt
}

// 如何使用默认值
// 使用default标签为字段定义默认值
/*
func main() {
	db, err := gorm.Open(
		mysql.Open("root:103003@tcp(127.0.0.1:3306)/gorm?charset=utf8mb4&parseTime=True&loc=Local"),
		&gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//Create  创建数据
	db.Create(&Product{Code: "042", Price: 100})
	// Read	查询数据
	var product Product
	db.First(&product, 1)              //根据整形主键查找
	db.First(&product, "code=?", "042") // 查找code 字段为 042 的记录
	//First 只支持查询一条记录
	//Find  查询多条记录

	//Updata 更新数据 将product 的 price 更新为 200
	db.Model(&product).Update("Price", 200)
	//Update -- 更新多个字段
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})
	// Delete -- 删除product
	db.Delete(&product, 1)

}*/

func main() {
	//gorm 性能优化
	db, err := gorm.Open(mysql.Open("root:103003@tcp(127.0.0.1:3306)gorm?charset=utf8mb4"),
		&gorm.Config{
			SkipDefaultTransaction: true, //关闭默认事务
			PrepareStmt:            true, //缓存预编译语句
		})
	if err != nil {
		panic("failed to connect database")
	}
	//Gorm 创建数据
	//创建一条
	p := &Product{Code: "042"}
	res := db.Create(p)
	fmt.Println(res.Error) //获取error
	fmt.Println(p.ID)      //返回插入数据的主键
	//创建多条数据
	products := []*Product{{Code: "041"}, {Code: "042"}, {Code: "043"}}
	res = db.Create(products)
	fmt.Println(res.Error)
	for _, p := range products {
		fmt.Println(p.ID)
	}
	//如何使用Upset
	// 以不处理冲突为例，创建一条数据
	//p = &Product{Code:"D42",ID:1}
	//使用clause.OnConflict处理数据冲突
	//db.Clauses(clause.OnConflict{DoNothing:true}).Create(&p)

	//Gorm 查询数据
	//获取第一条记录(主键升序),查询不到数据则返回ErrRecordNotFound
	u := &User{}
	db.First(u) //SELECT * FROM users ORDER BY id LIMIT 1
	//查询多条数据
	users := make([]*User, 0)
	result := db.Where("age > 10 ").Find(&users) //SELECT * FROM user WHERE age > 10
	fmt.Println(result.RowsAffected)             //返回找到的记录数
	fmt.Println(result.Error)                    //return err
	db.Where("name IN ?", []string{"jinzhu", "jinzhu2"}).Find(&users)
	// IN SELECT * FROM user WHERE name IN ('jinzhu'.'jinzhu2')
	db.Where("name LIKE ?", "%jin%").Find(&users)
	// LIKE SELECT * FROM user WHERE name LIKE '%jin%'
	db.Where("name = ? AND age = ? ", "jinzhu", "18").Find(&users)
	// AND SELECT * FROM user WHERE name = "jinzhu" AND age >= 18
	db.Where(&User{Name: "jinzhu", Age: 0})
	//SELECT * FROM user WHERE name = "jinzhu"
	db.Where(map[string]interface{}{"Name": "jinzhu", "age": "18"})
	//SELECT * FROM user WHERE name = "jinzhu" AND age = 18

	//First的使用踩坑
	//使用First时，需要注意查询不到数据会返回ErrRecordNotFound
	//使用Find查询多条数据，查询不到数据不会返回错误
	// 使用结构体作为查询条件
	// 当使用结构体作为查询条件时，GORM只会查询非零值字段。这意味着如果你的字段为0，”false“或者其它零值，
	// 该字段不会被用于构建查询条件，使用MAP来构建查询条件。

	//条件更新单个列
	db.Model(&User{ID: 111}).Where("age > ?", 19).Update("nmae", "hello")
	// UPDATE users SET name = "hello" WHERE age > 19

	// 更新多个列
	//使用struct更新时，只会更新非零值.
	//如果需要更新零值可以使用MAP更新或者使用select选择字段
	db.Model(&User{ID: 111}).Updates(User{Name: "hello", Age: 19})
	//UPDATE users SET name = "hello",age = 19 WHERE id = 111
	db.Model(&User{ID: 111}).Updates(map[string]interface{}{"name": "hello", "age": 19, "actived": false})
	//更新选定字段
	db.Model(&User{ID: 111}).Select("name").Updates(map[string]interface{}{"name": "hello", "age": 19, "actived": false})
	//SQL 表达式更新
	db.Model(&Product{ID: 3}).Update("age", gorm.Expr("age * ? + ? ", 2, 100))
	//UPDATE "products" SET "price"=price*2+100 WHERE "id" = 3

	//GORM删除数据
	//物理删除
	db.Delete(&User{}, 10)
	db.Delete(&User{}, "10")
	db.Delete(&User{}, []int{1, 2, 3})
	//DELETE FROM user WHERE id = 10
	db.Where("name LIKE", "%jinzhu%").Delete(User{})
	db.Delete(User{}, "name LIKE ? ", "%inzhu%")
	//DELETE FROM user WHERE name LIKE "%jinzhu%"

	//软删除
	//删除一条
	u = &User{ID: 111}
	db.Delete(&u)
	//批量删除
	db.Where("age =?", 20).Delete(&User{})
	db.Where("age =20").Find(&users)
	db.Unscoped().Where("age = 20").Find(&users)
	// 拥有软删除能力的Model调用Delete时，记录不会从数据库中真正删除。
	// 但GORM会将deletedAt置为当前时间，并且你不能通过正常的查询方法找到该记录
	// 使用Unscoped可以查询到被软删除的数据

	//GORM 事务
	tx := db.Begin() //开始事务

	if err = tx.Create(&User{Name: "name"}).Error; err != nil {
		tx.Rollback()
		//遇到错误时回滚事务
		return
	}
	if err = tx.Create(&User{Name: "name1"}).Error; err != nil {
		tx.Rollback()
		return
	}
	//提交事务
	tx.Commit()

	//Transaction方法用于自动提交事务，避免用户漏写Commit、Rollback
	if err = db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(&User{Name: "name"}).Error; err != nil {
			return err
		}
		if err = tx.Create(&User{Name: "name1"}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	}); err != nil {
		return
	}

}

// Gorm的约定
// Gorm使用名为ID的字段作为主键
// 使用结构体的蛇形负数作为表名
// 字段名的蛇形作为列名
// 使用CreatedAt、UpdatedAt字段作为创建、更新时间
//https://gorm.cn/docs/#Install
