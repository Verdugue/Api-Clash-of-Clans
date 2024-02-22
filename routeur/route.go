package route

import (
	"fmt"
	"net/http"
	"os"
	controller "pokemon/controller"
)

func InitRoute() {
	http.HandleFunc("/", controller.Index)
	http.HandleFunc("/search", controller.SearchPokemon)

	rootDoc, _ := os.Getwd()
	fileserver := http.FileServer(http.Dir(rootDoc + "/assets"))
	http.Handle("/static/", http.StripPrefix("/static/", fileserver))

	fmt.Println("(http://localhost:8080/) - Server started on port:8080")
	http.ListenAndServe("localhost:8080", nil)
	fmt.Println("Server closed")
}
