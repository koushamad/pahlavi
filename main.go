package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	router := gin.Default()

	router.SetFuncMap(template.FuncMap{
		"upper": strings.ToUpper,
	})

	router.GET("/", func(c *gin.Context) {
		resp, err := http.Get(os.Getenv("WEBSITE"))
		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatalln(err)
		}

		for key, headers := range resp.Header {
			for _, h := range headers {
				c.Header(key, h)
			}
		}

		content := replace(body)

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), content)
	})

	router.Any("/:name/*action", func(c *gin.Context) {
		resp := send(c)
		var content []byte

		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Fatalln(err)
		} else {
			content = replace(body)
		}

		for key, headers := range resp.Header {
			for _, h := range headers {
				c.Header(key, h)
			}
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), content)
	})

	router.Run("localhost:8080")
}

func replace(body []byte) []byte {
	content := strings.Replace(string(body), os.Getenv("SUB_DOMAIN_STATIC"), os.Getenv("FORWARD_SUB_DOMAIN_STATIC"), -1)
	content1 := strings.Replace(content, os.Getenv("BASE_DOMAIN"), os.Getenv("FORWARD_DOMAIN"), -1)

	return []byte(content1)
}

func revert(body string) string {
	content := strings.Replace(body, os.Getenv("FORWARD_SUB_DOMAIN_STATIC"), os.Getenv("SUB_DOMAIN_STATIC"), -1)
	content1 := strings.Replace(content, os.Getenv("FORWARD_DOMAIN"), os.Getenv("BASE_DOMAIN"), -1)

	return content1
}

func send(c *gin.Context) *http.Response {
	url := c.Request.URL.Path
	method := c.Request.Method
	name := c.Param("name")
	var req *http.Request
	var err error

	if name == "static" {
		req, err = http.NewRequest(method, os.Getenv("SUB_DOMAIN_STATIC")+strings.Replace(url, "/static", "", 1), c.Request.Body)
	} else {
		req, err = http.NewRequest(method, os.Getenv("BASE_DOMAIN")+url, c.Request.Body)
	}

	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	headers := http.Header{}

	for key, hs := range c.Request.Header {
		for _, h := range hs {
			headers.Add(key, revert(h))
		}
	}

	req.Header = headers

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	return resp
}
