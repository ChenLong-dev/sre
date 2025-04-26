package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/library/base/ctime"
)

func ExampleMiddleware() {
	router := gin.New()

	router.Use(ParseUserAgentMiddleware())

	router.Use(GetUUIDMiddleware("QT-User-Id"))

	router.Use(CatchPanicMiddleware())

	router.Use(TimeoutMiddleware(ctime.Duration(time.Second)))
}
