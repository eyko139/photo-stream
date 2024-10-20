package main


import ( 
	"github.com/julienschmidt/httprouter"
    "net/http"
)

func Routes() http.Handler {
    router := httprouter.New()
    // router.GET("/", Ren())
    return router
        
}

