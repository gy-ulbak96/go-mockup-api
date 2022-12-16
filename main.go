package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type server struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

var servers = make(map[string]server)


func main() {
	router := gin.Default()
	router.POST("/servers", postServer)
	router.GET("/servers/:uuid", getServer)
	router.DELETE("/servers/:uuid", deleteServer)
	router.GET("/servers/list", getServerlist)
	router.Run("localhost:8080")
}

func postServer(c *gin.Context) {
	var newServer server
	body := c.Request.Body
	value, err := io.ReadAll(body)
	if err != nil {
		log.Printf("%+v", err.Error())
		return

	}
	var data map[string]interface{}
	json.Unmarshal([]byte(value), &data)
	vmName, exist := data["name"].(string)
	if !exist {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "You have to specify VM Name"})
		return
	}
	min, max := 10000, 99999
	vmId := vmName + "-" + strconv.Itoa(rand.Intn(max-min)+min)
	newServer.Id = vmId
	newServer.Name = vmName
	servers[vmId] = newServer
	c.IndentedJSON(http.StatusCreated, newServer)
}

func getServer(c *gin.Context) {
	id := c.Param("uuid")
	for key, val := range servers {
		if key == id {
			c.IndentedJSON(http.StatusOK, val)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "server not found."})
}

func getServerlist(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, servers)
}

func deleteServer(c *gin.Context) {
	id := c.Param("uuid")
	for key, val := range servers {
		log.Printf("%s %s", key, val)
		if key == id {
			delete(servers, key)
			c.IndentedJSON(http.StatusNoContent, nil)
			return
		}
	}
	c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}
