package main

import "filestore-server/route"

func main() {
	r := route.Router()
	r.Run(":8080")
}
