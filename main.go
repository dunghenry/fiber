package main

import (
	"database/sql"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

type User struct {
	Name     string `json:"name" form:"name"`
	Id       int    `json:"id" form:"id"`
	Email    string `json:"email" form:"email"`
	Mobile   string `json:"mobile" form:"mobile"`
	Password string `json:"password" form:"password"`
}

func main() {
	app := fiber.New()
	// Run local vs mysql
	db, err := sql.Open("mysql", "root:@/mydb")
	// Run docker vs container
	// db, err := sql.Open("mysql", "root:root@tcp(database:3306)/fiber")
	if err != nil {
		log.Fatal(err)
	}
	app.Get("/", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, email, mobile, password FROM users")
		if err != nil {
			log.Fatal(err)
		}

		var users []User
		for rows.Next() {
			var user User
			rows.Scan(&user.Id, &user.Name, &user.Email, &user.Mobile, &user.Password)
			users = append(users, user)
		}
		defer rows.Close()
		return c.JSON(&fiber.Map{
			"success": true,
			"users":   users,
		})
	})

	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var user User
		row, err := db.Query("SELECT  id, name, email, mobile, password FROM users WHERE id = ?", id)
		if err != nil {
			log.Fatal(err)
		}
		for row.Next() {
			row.Scan(&user.Id, &user.Name, &user.Email, &user.Mobile, &user.Password)
		}
		defer row.Close()
		if user.Id == 0 {
			return c.JSON(&fiber.Map{
				"success": true,
				"msg":     "User not found",
			})
		}
		return c.JSON(&fiber.Map{
			"success": true,
			"users":   user,
		})
	})

	app.Post("/", func(c *fiber.Ctx) error {
		u := new(User)
		if err := c.BodyParser(u); err != nil {
			log.Fatal(err)
			return c.Status(400).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}
		stmt, err := db.Prepare("INSERT INTO users( name, email, mobile, password) VALUES (?, ?, ?, ?)")
		if err != nil {
			return c.Status(500).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}
		rs, err := stmt.Exec(u.Name, u.Email, u.Mobile, u.Password)
		if err != nil {
			log.Fatal(err)
		}
		id, err := rs.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		u.Id = int(id)
		defer stmt.Close()
		return c.JSON(&fiber.Map{
			"success": true,
			"user":    u,
		})
	})

	app.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		stmt, err := db.Prepare("DELETE FROM users WHERE id =?")
		if err != nil {
			log.Fatal(err)
		}
		rs, err := stmt.Exec(id)
		if err != nil {
			log.Fatalln(err)
		}
		row, err := rs.RowsAffected()
		if err != nil {
			log.Fatalln(err)
		}
		defer stmt.Close()
		rows := int(row)
		if rows > 0 {
			return c.JSON(&fiber.Map{
				"success": true,
				"count":   rows,
			})
		}
		return c.JSON(&fiber.Map{
			"success": true,
			"msg":     "User not found",
		})
	})
	app.Put("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		u := new(User)
		if err := c.BodyParser(u); err != nil {
			log.Fatal(err)
			return c.Status(400).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}
		stmt, err := db.Prepare("UPDATE users SET name=?, email=?, mobile=?, password=? WHERE id=?")
		if err != nil {
			log.Fatal(err)
		}
		rs, err := stmt.Exec(u.Name, u.Email, u.Mobile, u.Password, id)
		if err != nil {
			log.Fatal(err)
		}

		row, err := rs.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		i, _ := strconv.Atoi(id)
		u.Id = i
		rows := int(row)
		defer stmt.Close()
		if rows > 0 {
			return c.JSON(&fiber.Map{
				"success": true,
				"count":   u,
			})
		}
		return c.JSON(&fiber.Map{
			"success": true,
			"msg":     "User not found",
		})
	})
	app.Listen(":3000")
}
