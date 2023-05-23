package run

import (
	"gitlab.arc.hcloud.io/ccp/hcloud-platform/go-helm-api/src/server"
	"gitlab.arc.hcloud.io/ccp/hcloud-platform/go-helm-api/src/lb"
)

func Run() {
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