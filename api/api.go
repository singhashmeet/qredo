package api

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/singhashmeet/temp/pkg/jsonparser"
)

type API struct {
	host string
	port string
}

func (a *API) Sum(c *gin.Context) {
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(jsonData))
	json := jsonparser.NewParser(string(jsonData))
	err = json.IsValid()
	fmt.Println(err)
	fmt.Println(json.Floating)
	fmt.Println(json.Integers)
	c.JSON(http.StatusOK, "")
}

func (a *API) addPaths(router *gin.Engine) {
	router.POST("/sum", a.Sum)
}

func (a *API) ListenAndServe() {
	router := gin.Default()
	a.addPaths(router)
	router.Run(net.JoinHostPort(a.host, a.port))
}

func NewServer(host, port string) *API {
	return &API{
		host: host,
		port: port,
	}
}
