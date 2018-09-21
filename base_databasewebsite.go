/*
// Copyright 2018 Sendhil Vel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

base_website.go
Date 		: 19/07/2018
Comment 	: This is seed file for creating any go website which connects to Postgres database.
Version 	: 1.0.9
by Sendhil Vel
*/

package main

/*
	imports necessary packages
*/
import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"
)

/*
	Defining necessary variables
*/
var tpl *template.Template
var err error
var r *gin.Engine
var rpcurl string
var db *sql.DB

/*
Users is the core object
*/
type Users struct {
	DBId     string `json:"userid,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
	IsActive bool   `json:"isactive,omitempty"`
	Password string `json:"password,omitempty"`
}

/*
render - renders the pages
*/
func render(c *gin.Context, data gin.H, templateName string) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		c.XML(http.StatusOK, data["payload"])
	default:
		c.HTML(http.StatusOK, templateName, data)
	}
}

/*
jsonresponse - This function return the response in a json format
*/
func jsonresponse(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "process executed successfully",
	})
}

/*
performLogin - gets the posted values for varibles username and password and check if the username/password combination is valid
*/
func performLogin(c *gin.Context) {
	/*
		Get the values from POST objects
	*/
	username := c.PostForm("username")
	password := c.PostForm("password")

	/*
		Checks the username and password variables valuesare not empty
	*/
	if len(username) == 0 || len(password) == 0 {
		err = errors.New("missing password and/or email")
		return
	}

	/*
		Call the actual function which checks and return error or information about user
		Based on status we are redirecting to necessary pages
		If error then redirecting to login page again with error messages
		If valid then redirecting to userprofile page which display user information.
	*/
	UserInfo, err := getUser(username, password)
	if err != nil {
		render(c, gin.H{
			"title":        "Login",
			"ErrorMessage": "Login Failed",
		}, "login.html")
	} else {
		render(c, gin.H{
			"title":    "User Profile",
			"UserInfo": UserInfo,
		}, "userprofile.html")
	}
}

/*
getUser - this checks the information stored in database and information passed.
			In case of valid information, user information is returned.
			In case of invalid information error is returned.
*/
func getUser(vUserName string, vPassword string) (userobj Users, err error) {
	/*
		Defining the variables
	*/
	var vDBId, vName, vEmail, vToken, vsqlPassword sql.NullString
	var vIsActive sql.NullBool

	/*
		creating a sql query using parameter
	*/
	sqlStmt := fmt.Sprintf(`SELECT id,Name,Email,Token,Is_Active,Password  FROM shard_1.users WHERE LOWER(Email)=lower('%s') and lower(password) = md5('%s')`, strings.ToLower(vUserName), vPassword)

	/*
		Executing the sql query
		In case of error, error information will be returned
		User object is returned in case credentials are valid
	*/
	err = db.QueryRow(sqlStmt).Scan(&vDBId, &vName, &vEmail, &vToken, &vIsActive, &vsqlPassword)
	if err != nil {
		fmt.Println(err)
		err = fmt.Errorf("unknown email : %s", err.Error())
		return
	}
	userobj.DBId = vDBId.String
	userobj.Name = vName.String
	userobj.Email = vEmail.String
	userobj.Token = vToken.String
	userobj.IsActive = vIsActive.Bool
	userobj.Password = ""
	return
}

/*
initializeRoutes - This will defines various routes and relavant information
*/
func initializeRoutes(port string) {
	/*
		All the urls will be mentioned and configured.
	*/
	/*
		url : /test
	*/
	r.GET("/test", showHomePage)
	/*
		url : /
	*/
	r.GET("/", showHomePage)
	/*
		Defining group route for users
	*/
	userRoutes := r.Group("/user")
	{
		/*
			url : /user/
		*/
		userRoutes.GET("/", showHomePage)
		/*
			url : /user/login (method is get)
		*/
		userRoutes.GET("/login", showLoginPage)
		/*
			url : /user/login (method is post)
		*/
		userRoutes.POST("/login", performLogin)
		/*
			url : /user/jsonresponse
		*/
		userRoutes.GET("/jsonresponse", jsonresponse)
	}
	fmt.Println("-------Starting server-------------")
}

/*
main - main function of the file
*/
func main() {
	/*
		Loads the env variables
	*/
	gotenv.Load()

	/*
		Get the port no from .env file.
		Convert string to int
		In case some error comes then process is stopped
	*/
	port := os.Getenv("WEBSITE_PORT")
	dbUser := os.Getenv("USER_DB_USER")
	dbPass := os.Getenv("USER_DB_PASS")
	dbName := os.Getenv("USER_DB_NAME")
	dbURL := os.Getenv("USER_DB_URL")

	/*
		Setting Gin parameter and folders
	*/
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r = gin.Default()
	r.LoadHTMLGlob("templetes/html/*")
	r.Static("/css", "templetes/css")
	r.Static("/js", "templetes/js")
	r.Static("/img", "templetes/img")
	r.Static("/fonts", "templetes/fonts")

	/*
		Calling function which will be connecting to Database
		In case of error, we are stopping the execution.

	*/
	err := initDBConnection(dbUser, dbPass, dbURL, dbName)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	fmt.Println("DB connected")

	/*
		calling function to setup the routes
	*/
	go initializeRoutes(port)

	/*
		Starting the server in the specified port
	*/
	fmt.Println("Web Portal is running on " + port)
	r.Run(port)
	fmt.Println("-------Started server-------------")
}

/*
showHomePage - this will display status of website
*/
func showHomePage(c *gin.Context) {
	c.JSON(200, gin.H{
		"Server": "Cool you are ready to start website in goLang",
	})
}

/*
showLoginPage - This will load and show login page with necessary parameters
*/
func showLoginPage(c *gin.Context) {
	render(c, gin.H{
		"title":        "Login",
		"ErrorMessage": "",
	}, "login.html")
}

/*
initDBConnection - This function connects to Postgres database
*/
func initDBConnection(dbUser, dbPass, dbURL, dbNAME string) (err error) {
	/*
		Variables defined here
	*/
	var user, pass, url, name string

	/*
		verify that all variables exists
	*/
	if len(dbUser) == 0 || len(dbURL) == 0 || len(dbPass) == 0 || len(dbNAME) == 0 {
		err = errors.New("Missing DB Credentails. Please Check")
		return
	}

	/*
		verify the varibles and set values after remove spaces
	*/
	if len(dbUser) > 0 && len(dbPass) > 0 && len(dbURL) > 0 && len(dbNAME) > 0 {
		user = strings.TrimSpace(dbUser)
		pass = strings.TrimSpace(dbPass)
		url = strings.TrimSpace(dbURL)
		name = strings.TrimSpace(dbNAME)
	}

	/*
		Prepares the connection string
	*/
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require", user, pass, url, name)
	fmt.Printf("connecting to database: %s\n", url)

	/*
		connects the database with the provided values, in case of any issue error will be raise
	*/
	db, err = sql.Open("postgres", connString)
	if err != nil {
		err = fmt.Errorf("Database refused connection: %s", err.Error())
		return
	}

	return
}
