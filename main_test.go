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
	// "time"
)

type server struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

type lb struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Protocol string `json:"protocol"`
	Vip string `json:"vip"`
	Endpoint Endpoint `json:"endpoint,omitempty"`
}

type Endpoint struct {
  Id string `json:"id"`
	Server []server `json:"server"`
	Lbport int `json:"lbport"`
	Serverport int `json:"serverport"`
}

var servers = make(map[string]server)
var lbs = make(map[string]lb)


func main() {
	router := gin.Default()
	router.POST("/servers", postServer)
	router.GET("/servers", getServerlist)
	router.GET("/servers/:uuid", getServer)
	router.DELETE("/servers/:uuid", deleteServer)
	

	router.POST("/lbs", postLB)
	// router.GET("/lbs", getLBlist)
	// router.GET("/lbs/:uuid", getLB)
	// router.POST("/lbs/:uuid", postLBBind)
	// router.POST("/lbs/:uuid/update", updateLBBind)
	// router.DELETE("/lbs/:uuid", deleteLB)
	// router.Run("localhost:8080")
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
		//invalid operation: cannot slice lbs (variable of type map[string]lb) > cuz map has no order
		recentlb := lbs[:len(lbs)-1]
	  lbVip, _ = NextIP(recentlb.Vip)
	}
	newLB.Id = lbId
	newLB.Name = lbName
	newLB.Protocol = lbProtocol
	newLB.Vip = lbVip
	lbs[lbId]= newLB
	c.IndentedJSON(http.StatusCreated, newLB)
	
}

func NextIP(ip_whole string) (string, error) {
	ip_parts_s := strings.Split(ip_whole, ".")
	//num string to int
	ip_parts := make([]int, 4)
  for i, v := range ip_parts_s {
    ip_parts[i], _ = strconv.Atoi(v)
  }
	for i, _ := range ip_parts {
    if ip_parts[3-i] == 255 {
			if ip_parts[2-i] == 255{
				continue
			}	else {
				ip_parts[2-i] += 1
				j := 3-i
				for j < 4 {
					ip_parts[j] = 0
					j += 1
				}
			}
		}
	}
	for i, v := range ip_parts {
    ip_parts_s[i] = strconv.Itoa(v)
  }
  ip_whole = strings.Join(ip_parts_s, ".")
	return ip_whole, nil
}

// func getLBlist(c *gin.Context) {

// }

// func getLB(c *gin.Context) {

// }

// func postLBBind(c *gin.Context) {

// }

// func updateLBBind(c *gin.Context) {

// }

// func deleteLB(c *gin.Context) {

// }

