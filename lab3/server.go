package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type DataStore struct {
	Data map[int]string
}

func (d *DataStore) Get(key int) (string, error) {
	if val, ok := d.Data[key]; ok {
		// the key exists in the data store.
		// return the value associated with the key.
		return val, nil
	}

	// the data store does not have the key.
	// return an error with an appropriate message.
	return "", errors.New(fmt.Sprintf("Key %d not found.", key))
}

func (d *DataStore) Set(key int, val string) error {
	// set the val for the key in the data store.
	d.Data[key] = val

	// no error to return
	return nil
}

func (d *DataStore) UnSet(key int) error {
	// unset the key in the data store.
	// ignore if the key doesn't exist
	delete(d.Data, key)

	// no error to return
	return nil
}

func NewDataStore() *DataStore {
	// create the data store
	ds := DataStore{}
	// allocate memory for the data store
	ds.Data = make(map[int]string)

	return &ds
}

type HTTPServer struct {
	Ports []int
}

func (h *HTTPServer) Start() {
	// create a done channel to signal server shutdown
	done := make(chan bool)

	// define a goroutine to start the server and serve requests
	startAndServe := func(index int) {
		go func() {
			fmt.Println("starting server at port:", h.Ports[index])

			// create a data store to be used by this server
			dataStore := NewDataStore()

			// define a server mux to add handlers
			mux := http.NewServeMux()

			// define url pattern regex
			getAllPattern := regexp.MustCompile(`^/$`)
			getOnePattern := regexp.MustCompile(`^/([0-9]+)$`)
			putOnePattern := regexp.MustCompile(`^/([0-9]+)/([0-9a-zA-Z]+)$`)

			// attach "/" match all route
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// match the routes based on the pattern
				switch {
				case getAllPattern.MatchString(r.URL.Path):
					if r.Method == "GET" {
						// create a slice with map (for 10 objects to start with)
						// where map's key is string and value is either int or string (interface{} type)
						response := make([]map[string]interface{}, 0)

						// add the data from the data store to the response slice
						for key, val := range dataStore.Data {
							response = append(response, map[string]interface{}{
								"key":   key,
								"value": val,
							})
						}

						// convert to JSON (ignore the error for now)
						jsonResponse, _ := json.Marshal(response)

						// update the writer with response
						w.WriteHeader(200)
						w.Write(jsonResponse)
					} else {
						// no other method allowed on this route
						w.WriteHeader(405)
					}
				case getOnePattern.MatchString(r.URL.Path):
					if r.Method == "GET" {
						// get the key from url
						matches := getOnePattern.FindAllStringSubmatch(r.URL.Path, -1)

						// extract the key from matches
						// and convert it to integer
						// ignore the error for now
						// [['/1', '1']] --> matches[0]: ['/1', '1'] --> matches[0][1]: '1'
						key, _ := strconv.Atoi(matches[0][1])

						// get the value from the data store
						val, err := dataStore.Get(key)
						if err != nil {
							w.WriteHeader(404)
						} else {
							// generate the json response using the key and value
							jsonResponse, _ := json.Marshal(map[string]interface{}{
								"key":   key,
								"value": val,
							})

							// update the writer with response
							w.WriteHeader(200)
							w.Write(jsonResponse)
						}
					} else {
						// no other method allowed on this route
						w.WriteHeader(405)
					}
				case putOnePattern.MatchString(r.URL.Path):
					if r.Method == "PUT" {
						// get the key and from url
						matches := putOnePattern.FindAllStringSubmatch(r.URL.Path, -1)

						// extract key and value from the matches
						key, _ := strconv.Atoi(matches[0][1])
						val := matches[0][2]

						// add the key and value to the data store
						dataStore.Set(key, val)

						w.WriteHeader(204)
					} else {
						// no other method allowed on this route
						w.WriteHeader(405)
					}
				default:
					// no such route found
					w.WriteHeader(404)
				}
			})

			// listen and serve http requests
			http.ListenAndServe(fmt.Sprintf(":%d", h.Ports[index]), mux)

			// signal the goroutine end
			fmt.Println("shutting down server at port:", h.Ports[index])
			done <- true
		}()
	}

	// start all the servers one by one
	for i := 0; i < len(h.Ports); i++ {
		// start a server and start serving requests in background
		startAndServe(i)
	}

	// wait for all the goroutines to end
	for i := 0; i < len(h.Ports); i++ {
		// receive the signal
		<-done
	}
}

func NewHTTPServer(ports []int) *HTTPServer {
	// create a http server instance
	hs := HTTPServer{}

	// assign the ports
	hs.Ports = ports

	return &hs
}

func main() {
	// generate port numbers
	if len(os.Args) < 2 {
		fmt.Println("usage: go run server.go 8001-8005")
		os.Exit(1)
	}

	// get the start and end ports
	startEndPort := strings.Split(os.Args[1], "-")
	startPort, _ := strconv.Atoi(startEndPort[0])
	endPort, _ := strconv.Atoi(startEndPort[1])

	// create a ports array to store all the ports
	ports := make([]int, 0)
	for i := startPort; i <= endPort; i++ {
		ports = append(ports, i)
	}

	// create a http server instance
	// with the required number of servers
	server := NewHTTPServer(ports)

	// start all the servers
	server.Start()
}