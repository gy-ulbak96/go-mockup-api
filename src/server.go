package run

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"fmt"
)

type server struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Ip string `json:"ip"`
}

type Endpoint struct {
	Server []string `json:"server"`
	Serverport int `json:"serverport"`
}

var servers = make(map[string]server)
var servers_recent []string

// About Backend Server
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
	serverip := ""
	if len(servers) == 0 {
		serverip = "172.0.0.1"
	}	else {
		recentserver := servers_recent[len(servers)-1]
	  serverip, err = NextIP(recentserver)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "There is no allocatable IP"})
			return
		}
	}
	newServer.Ip = serverip
	servers[vmId] = newServer
	servers_recent = append(servers_recent, serverip)
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
	c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "There is no such Server"})
}

func NextIP(ip_whole string) (string, error) {
	ip_parts_s := strings.Split(ip_whole, ".")
	//num string to int
	ip_parts := make([]int, 4)
  for i, v := range ip_parts_s {
    ip_parts[i], _ = strconv.Atoi(v)
  }
	for i := len(ip_parts) - 1; i >= 0; i--{
		if ip_parts[i] < 255 {
      ip_parts[i] += 1
			break
		}
    ip_parts[i] = 0
		if i == 0 {
			return "",fmt.Errorf("an error occurred")
		}
    ip_parts[i-1] += 1	
		if ip_parts[i-1] < 255{
			break
		}
		
	}
	for i, v := range ip_parts {
  	ip_parts_s[i] = strconv.Itoa(v)
  }
  ip_whole = strings.Join(ip_parts_s, ".")
	return ip_whole, nil
}