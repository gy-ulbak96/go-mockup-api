package main

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

type lb struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Protocol string `json:"protocol"`
	Vip string `json:"vip"`
	Endpoint Endpoint `json:"endpoint,omitempty"`
}

type Endpoint struct {
	Server []string `json:"server"`
	Lbport int `json:"lbport"`
	Serverport int `json:"serverport"`
}

var servers = make(map[string]server)
var lbs = make(map[string]lb)
var lbs_recent []string
var servers_recent []string


func main() {
	router := gin.Default()
	router.POST("/servers", postServer)
	router.GET("/servers", getServerlist)
	router.GET("/servers/:uuid", getServer)
	router.DELETE("/servers/:uuid", deleteServer)
	

	router.POST("/lbs", postLB)
	router.GET("/lbs", getLBlist)
	router.GET("/lbs/:uuid", getLB)
	router.POST("/lbs/:uuid", postLBBind)
	router.DELETE("/lbs/:uuid", deleteLB)
	router.Run("localhost:8080")
}

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

// About LB
func postLB(c *gin.Context) {
  var newLB lb
	body := c.Request.Body
	value, err := io.ReadAll(body)
	if err != nil {
		log.Printf("%+v", err.Error())
		return
	}
  var data map[string]interface{}
  json.Unmarshal([]byte(value), &data)
  lbName, exist := data["name"].(string)
	if !exist {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "You have to specify LB Name"})
		return
	}
	lbProtocol, exist := data["protocol"].(string)
	if !exist {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "You have to specify LB Protocol"})
		return
	}
	min, max := 10000, 99999
	lbId := lbName + "-" + strconv.Itoa(rand.Intn(max-min)+min)
	// Make Next LB VIP that increase 1
	lbVip := ""
	if len(lbs) == 0 {
		lbVip = "192.0.0.1"
	}	else {
		recentlb := lbs_recent[len(lbs)-1]
	  lbVip, err = NextIP(recentlb)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "There is no allocatable IP"})
			return
		}
	}
	newLB.Id = lbId
	newLB.Name = lbName
	newLB.Protocol = lbProtocol
	newLB.Vip = lbVip
	lbs[lbId]= newLB
	lbs_recent = append(lbs_recent, lbVip)
	log.Printf("%v",lbs_recent, )
	c.IndentedJSON(http.StatusCreated, newLB)
	
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


func getLBlist(c *gin.Context) {
  c.IndentedJSON(http.StatusOK, lbs)
}

func getLB(c *gin.Context) {
  id := c.Param("uuid")
	for key, val := range lbs {
		if key == id {
			c.IndentedJSON(http.StatusOK, val)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "lb not found."})
}

func postLBBind(c *gin.Context) {
	body := c.Request.Body
	value, err := io.ReadAll(body)
	if err != nil {
		log.Printf("%+v", err.Error())
		return
	}

	id := c.Param("uuid")
	var data map[string]interface{}
	json.Unmarshal([]byte(value), &data)
	

	var serverlist []string
	for _, element := range data["serverlist"].([]interface{}){
		serverlist = append(serverlist,fmt.Sprintf("%v", element))
	}
	lbport,_ := strconv.Atoi(data["lbport"].(string))
	serverport,_ := strconv.Atoi(data["serverport"].(string))
	  
	for _,serverval := range serverlist {
		if len(servers) == 0{
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "You didn't make any servers"})
			return
		}
		exists := false
    for _, server := range servers {
        if server.Ip == serverval {
            exists = true
            break
        }
    }
    if !exists {
        message := fmt.Sprintf("You didn't make this server %v", serverval)
        c.IndentedJSON(http.StatusNotFound, gin.H{"message": message})
        return
    }
	}

	if len(lbs) == 0{
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "You didn't make any lbs"})
		return
	}

	exists := false
	for key, val := range lbs {
		if key == id {
			exists = true
			log.Printf("%T, %v, %T, %v",serverlist, serverlist ,lbport, lbport)
			val.Endpoint.Server = serverlist
			val.Endpoint.Lbport = lbport
			val.Endpoint.Serverport = serverport
			lbs[key] = val
			c.IndentedJSON(http.StatusOK, val)
		} 
	}
	if !exists {
		message := fmt.Sprintf("LB that have such uuid %v don't exist", id)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": message})
		return
	}
	

	c.JSON(http.StatusOK, gin.H{
	"serverlist": data["serverlist"],
	"lbport": data["lbport"],
	"serverport": data["serverport"],
	})
  
}

	

func deleteLB(c *gin.Context) {
	id := c.Param("uuid")
	for key, val := range lbs {
		log.Printf("%s %s", key, val)
		if key == id {
			delete(lbs, key)
			c.IndentedJSON(http.StatusNoContent, nil)
			return
		}
	}
	c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}

