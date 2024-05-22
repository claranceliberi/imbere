package utils

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func ReturnError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": message,
	})
	debug.PrintStack()
	fmt.Println("Error occured %+v\n", message)

}
