package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/gin-gonic/gin"
)

var APPID, APP_CERTIFICATE string 

func main() {
	fmt.Println("Agora Token Builder")
	os.Setenv("APP_ID", "18aa7610b5a94be68a09484435b3e780")
	os.Setenv("APP_CERTIFICATE", "23f2f14910b2499a980ecaf579ff61de")

	appIDEnv, appIDExists := os.LookupEnv("APP_ID")
	appCertEnv, appCertExists := os.LookupEnv("APP_CERTIFICATE")
	 
	if !appIDExists || !appCertExists {
		log.Fatal("FATAL ERROR : ENV not properly configured, check APP_ID and APP_CERTIFICATE")
	} else {
		APPID = appIDEnv
		APP_CERTIFICATE = appCertEnv
	}
// Initialize the Gin router with the prefix
api := gin.Default()
api.Use(gin.Logger())
api.Use(gin.Recovery())

// Add the API prefix
apiGroup := api.Group("/api")

// Define a basic ping endpoint
apiGroup.GET("/ping", func(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
})

// Define API endpoints for generating Agora RTC and RTM tokens
apiGroup.GET("/rtc/:channelName/:role/:tokenType/:uid", getRtcToken)
apiGroup.GET("/rtm/:uid", getRtmToken)
apiGroup.GET("/rte/:channelName/:role/:tokenType/:uid", getBothTokens)

	api.Run("0.0.0.0:8000")
}


func getRtcToken(c *gin.Context)  {
	//get param values
	 channelName,tokenType, uidStr, role, expireTimeStamp, err := parseRTCParams(c)

	 if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"message":"Error Generating RTC Token: " + err.Error(),
			"status": 400,
		})
		return
	 }

	// generate the token
	rtcToken, tokenErr := generateRTCToken(channelName,uidStr,tokenType,role,expireTimeStamp)

	//return the token in JSON response
	if tokenErr != nil {
		log.Println(tokenErr)
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"message": "Error Generating RTC token: " + tokenErr.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"rtcToken":rtcToken,
		})
	}
}

func getRtmToken(c *gin.Context)  {
	//get param values
	uidStr, expireTimeStamp, err := parseRTMParams(c)

	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"message": "Error Generating rtm token",
		})
		return
	}

	//build rtm token
	rtmToken, tokenErr := rtmtokenbuilder.BuildToken(APPID,APP_CERTIFICATE, uidStr, rtmtokenbuilder.RoleRtmUser, expireTimeStamp)

	// return rtm token

	if tokenErr != nil {
		log.Println(tokenErr)
		c.Error(tokenErr)
		errMsg := "Error Generating RTM Token : " + tokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status":400,
			"error": errMsg,
		})
	} else {
		c.JSON(200, gin.H{
			"rtmToken": rtmToken,
		})
	}
}
func getBothTokens(c *gin.Context)  {
	//get the params
	channelName, tokenType, uidStr, role, expireTimeStamp, rtcParamErr := parseRTCParams(c)
	if rtcParamErr != nil {
		c.Error(rtcParamErr)
		c.AbortWithStatusJSON(400, gin.H{
			"status":400,
			"message": "Error generating tokens : " + rtcParamErr.Error(),
		})
	} 
	//generate rtc token
	rtcToken, rtcTokenErr := generateRTCToken(channelName, uidStr, tokenType,role,expireTimeStamp)
	//generate rtm token
	rtmToken, rtmTokenErr := rtmtokenbuilder.BuildToken(APPID,APP_CERTIFICATE,uidStr,rtmtokenbuilder.RoleRtmUser, expireTimeStamp)
	
	//return both tokens
	if rtcTokenErr != nil {
		c.Error(rtcTokenErr)
		errMsg := "Error generating RTC Token: "+ rtcTokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"message": errMsg,
		})
	} else if rtmTokenErr != nil {
		c.Error(rtmTokenErr)
		errMsg := "Error generating RTM Token: "+ rtmTokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"message": errMsg,
		})
	} else {
		c.JSON(200, gin.H{
			"rtcToken": rtcToken,
			"rtmToken": rtmToken,
		})
	}

}


func parseRTCParams(c *gin.Context) (channelName, tokenType,uidStr string, role rtctokenbuilder.Role, expireTimeStamp uint32, err error )  {
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokenType = c.Param("tokenType")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry","3600")

	if roleStr == "publisher" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10 ,64)

	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing error %s", expireTime, parseErr)
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimeStamp := uint32(time.Now().UTC().Unix())
	expireTimeStamp = currentTimeStamp + expireTimeInSeconds

	return channelName, tokenType, uidStr, role,expireTimeStamp, err
}

func parseRTMParams(c *gin.Context) (uidStr string, expireTimeStamp uint32, err error )  {
	//get param values
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry","3600")
	 
	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)

	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing error: %s",expireTime, parseErr)
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimeStamp := uint32(time.Now().UTC().Unix())
	expireTimeStamp = currentTimeStamp + expireTimeInSeconds

	return uidStr, expireTimeStamp, err

}

func generateRTCToken(channelName,uidStr, tokenType string, role rtctokenbuilder.Role, expireTimeStamp uint32 ) (rtcToken string, err error) {
	//check token type 
	if tokenType == "userAccount" {
		rtcToken, err := rtctokenbuilder.BuildTokenWithUserAccount(APPID, APP_CERTIFICATE, channelName, uidStr, role, expireTimeStamp)

		return rtcToken, err
	} else if tokenType == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)

		if parseErr != nil {
				err = fmt.Errorf("failed to parse uidStr: %s, causing error %s", uidStr, parseErr)
				return "", err
		}
		uid := uint32(uid64)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUID(APPID, APP_CERTIFICATE, channelName, uid, role, expireTimeStamp)
		return rtcToken, err
	} else {
		err = fmt.Errorf("failed to generate RTC token for unknown tokenType: %s, causing error %s", tokenType)
		log.Println(err)
		return "",err
	}
}