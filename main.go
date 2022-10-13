package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"math/rand"

	echo "github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var domainName = "http://localhost:1323/" // change it to your domainname:port

func checkIfDatabaseExists() {
	var err error
	DB, err = sql.Open("sqlite3", "./database.db")
	if os.IsNotExist(err) {
		_, err = os.Create("./database.db")
		if err != nil {
			fmt.Println(err)
		}
		DB, err = sql.Open("sqlite3", "./database.db")
	}
	_, err = DB.Exec("CREATE TABLE `url_table` (`uid` INTEGER PRIMARY KEY AUTOINCREMENT, `redirect_uri` LONGTEXT NULL, `server_slug` LONGTEXT NULL, `created` DATE NULL)")
	if err != nil {
		fmt.Println(err)
	}

}

func printAllDatabse() {
	query, err := DB.Query("SELECT * FROM url_table")
	if err != nil {
		fmt.Println(err)
	}
	for query.Next() {
		var uid int
		var username string
		var departname string
		var created string
		err = query.Scan(&uid, &username, &departname, &created)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(uid, username, departname, created)
	}
}

func createRedirection(redirect_uri string, server_slug string) string {
	// check if slug already exist in databse
	var row_count int
	err := DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM url_table WHERE server_slug='%s'", server_slug)).Scan(&row_count)
	if err != nil {
		fmt.Println(err)
	}
	if row_count > 0 {
		server_slug = fmt.Sprintf("%s-%d", server_slug, rand.Intn(99999))
		fmt.Println("Slug already exist, new slug is " + server_slug)
	}
	// insert
	insert, err := DB.Prepare("INSERT INTO url_table(redirect_uri, server_slug, created) values(?,?,?)")
	if err != nil {
		fmt.Println(err)
	}
	_, err = insert.Exec(redirect_uri, server_slug, time.Now())
	if err != nil {
		fmt.Println(err)
	}
	return server_slug
}

func main() {
	checkIfDatabaseExists()
	e := echo.New()
	e.GET("/:path", func(c echo.Context) error {
		path := c.Param("path")
		var redirect_uri string
		err := DB.QueryRow(fmt.Sprintf("SELECT redirect_uri FROM url_table WHERE server_slug='%s'", path)).Scan(&redirect_uri)
		if err != nil {
			fmt.Println(err)
		}
		return c.Redirect(301, redirect_uri)
	})
	e.POST("/", func(c echo.Context) error {
		redirect_uri := c.FormValue("redirect_uri")
		server_slug := c.FormValue("server_slug")
		created_slug := createRedirection(redirect_uri, server_slug)
		return c.String(200, domainName+created_slug)
	})
	printAllDatabse()
	e.Logger.Fatal(e.Start(":1323"))
}
