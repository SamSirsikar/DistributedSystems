package main

import (
	"fmt"
	"github.com/drone/routes"
	"log"
	"net"
	"net/http"
	"os"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"net/rpc"
	"github.com/mkilling/goejdb"
    "labix.org/v2/mgo/bson"
    "github.com/naoina/toml"
)


// define a application config
type Config struct {
	Database struct {
		FileName string
	}
	PortNum int
	Replication struct {
		RpcServerPortNum int
		Replica []string
	}
}

// define a user profile
type Profile struct {
	Email string `json:"email"`
	Zip string `json:"zip"`
	Country string `json:"country"`
	Profession string `json:"profession"`
	FavoriteColor string `json:"favorite_color"`
	IsSmoking string `json:"is_smoking"`
	FavoriteSport string `json:"favorite_sport"`
	Food struct {
		Type string `json:"type"`
		DrinkAlcohol string `json:"drink_alcohol"`
	} `json:"food"`
	Music struct {
		SpotifyUserId string `json:"spotify_user_id"`
	} `json:"music"`
	Movie struct {
		TvShows []string `json:"tv_shows"`
		Movies []string `json:"movies"`
	} `json:"movie"`
	Travel struct {
		Flight struct {
			Seat string `json:"seat"`
		} `json:"flight"`
	} `json:"travel"`
}

// define a profile manager
type ProfileManager struct {
	Data map[string]*Profile
	Conn *goejdb.Ejdb
	Coll *goejdb.EjColl
	Persist bool
	Clients []*rpc.Client
	Replicas []string
}


func (p *ProfileManager) Get(key string) *Profile {
	// reads data stored in the ProfileManager
	if p.Persist {
		query := fmt.Sprintf("{\"email\" : \"%s\"}", key)
		res, _ := p.Coll.Find(query)
		log.Printf("GET >> Records found: %d\n", len(res))
		if len(res) > 0 {
				profile := Profile{}
		        bson.Unmarshal(res[0], &profile)
		        return &profile
    	}
	} else {
		if profilePtr, ok := p.Data[key]; ok {
  			return profilePtr
		}	
	}
	
	return nil
}

func (p *ProfileManager) Set(key string, val *Profile, replicate bool) {
	// saves the data into the ProfileManager
	if p.Persist {
		// check if already exists and delete
		if p.Get(key)!=nil{
			p.UnSet(key, replicate)
		}
		bsrec, _ := bson.Marshal(val)
	    p.Coll.SaveBson(bsrec)
	    log.Println("SET >> Record saved")
	    
	    if replicate {
			var reply bool
			for i := range p.Clients {
				// check if the client is connected
				if p.Clients[i] == nil {
					client, ejErr := rpc.Dial("tcp", p.Replicas[i])
					if ejErr != nil {
						log.Println(ejErr)
						continue
					}
					p.Clients[i] = client	
				}
				log.Println("SET >> Set RPC initiated")
				err := p.Clients[i].Call("RPC.Set", RPCParams{Key: key, Val: val}, &reply)
				if err != nil {
					log.Fatal(err)
				}
			}
	    }
	} else {
		p.Data[key] = val
	}
}

func (p *ProfileManager) UnSet(key string, replicate bool) {
	// removes the data from ProfileManager based on the key
	if p.Persist{
		query := fmt.Sprintf("{\"email\" : \"%s\", \"$dropall\" : true }", key)
		fmt.Println(query)
		_, err := p.Coll.Update(query)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("UNSET >> Record deleted / updated")	
		}
		if replicate {
			var reply bool
			for i := range p.Clients {
				// check if the client is connected
				if p.Clients[i] == nil {
					client, ejErr := rpc.Dial("tcp", p.Replicas[i])
					if ejErr != nil {
						log.Println(ejErr)
						continue
					}
					p.Clients[i] = client	
				}
				log.Println("UNSET >> UnSet RPC initiated")
				err := p.Clients[i].Call("RPC.UnSet", RPCParams{Key: key, Val: nil}, &reply)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	} else {
		delete(p.Data, key)
	}
}


func New(config *Config) *ProfileManager {
	// creates and returns a new ProfileManager object
	pm := ProfileManager{}
	pm.Persist = true // hardcoded for now
	if pm.Persist {
		// create database connection
		jb, err := goejdb.Open(config.Database.FileName, goejdb.JBOWRITER | goejdb.JBOCREAT | goejdb.JBOTRUNC)
	    if err != nil {
	        os.Exit(1)
	    }
	    pm.Conn = jb
		coll, _ := jb.CreateColl("Profile", nil)
		pm.Coll = coll
		
		// create rpc connection
		replicaCount := len(config.Replication.Replica)
		clients := make([]*rpc.Client, replicaCount)
		log.Println("Trying to establish TCP connections")
		for i := 0; i < replicaCount; i++ {
    		client, ejErr := rpc.Dial("tcp", config.Replication.Replica[i])
			if ejErr != nil {
				log.Println(ejErr)
				continue
			}
			clients[i] = client
		}
		pm.Clients = clients
		pm.Replicas = config.Replication.Replica
	} else {
		pm.Data = make(map[string]*Profile)
	}
	
	return &pm
}

var profileManager *ProfileManager

func GetProfile(w http.ResponseWriter, r *http.Request) {
	// get email from the url
	params := r.URL.Query()
	email := params.Get(":email")
	
	// get the profile associated with the email
	profile := profileManager.Get(email)
	if profile == nil {
		// no such user found - return 404
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	// user found - serialize the user as json and return
	response, _ := json.Marshal(*profile)
	
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(response)
}

func PostProfile(w http.ResponseWriter, r *http.Request) {
	// create an empty profile
	profile := Profile{}
	
	// read profile data from the request body
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	requestBody := buf.Bytes()
	
	// parse the request body (json content) to the Profile struct
	json.Unmarshal(requestBody, &profile)
	
	// save the profile struct in the using the manager
	profileManager.Set(profile.Email, &profile, true)
    
    // return 200 status
	w.WriteHeader(http.StatusCreated)
}

func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	// get email from the url
	params := r.URL.Query()
	email := params.Get(":email")
	
	// delete the user corresponding to the email
	profileManager.UnSet(email, true)
	
	// return a 204
	w.WriteHeader(http.StatusNoContent)
}

func PutProfile(w http.ResponseWriter, r *http.Request) {
	// get email from the url
	params := r.URL.Query()
	email := params.Get(":email")
	
	// get the user corresponding to the email
	profile := profileManager.Get(email)
	if profile == nil {
		// no such user found - return 404
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	// read profile data from the request body 
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	requestBody := buf.Bytes()
	
	// parse the request body (json content) and update the Profile struct
	json.Unmarshal(requestBody, &profile)
	
	// save the updated profile back using the manager
    profileManager.Set(profile.Email, profile, true)
	
	// return a 204
	w.WriteHeader(http.StatusNoContent)
}

// define the RPC listener and the related remote procedures
type RPC int

type RPCParams struct {
	Key string
	Val *Profile
}

func (r *RPC) Set(params RPCParams, ack *bool) error {
	profileManager.Set(params.Key, params.Val, false)
	return nil
}

func (r *RPC) UnSet(params RPCParams, ack *bool) error {
	profileManager.UnSet(params.Key, false)
	return nil
}

func ListenAndServeRPC(config *Config) {
	// form the address to listen on
	address := fmt.Sprintf("0.0.0.0:%d", config.Replication.RpcServerPortNum)
	addy, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	// create a tcp connection on the configured address
	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	// register the listener to the rpc module
	listener := new(RPC)
	rpc.Register(listener)
	
	// start accepting inbound tcp connections
	log.Println("TCP server listening at:", address)
	rpc.Accept(inbound)
}

func main() {
	// check the arguments
	if len(os.Args) <= 1 {
		fmt.Println("Please provide the config file. Usage: go run app.go config.toml")
		os.Exit(1)
	}
	
	// get the config file name
	configFile := os.Args[1]
	
	// open the config file
	f, err := os.Open(configFile)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    
    // create a buffer to read the config
    buf, err := ioutil.ReadAll(f)
    if err != nil {
        panic(err)
    }
    
    // load the config file into the config struct
    log.Println("Loading config")
    var config Config
    if err := toml.Unmarshal(buf, &config); err != nil {
        panic(err)
    }
	
	// create a profile manager
	profileManager = New(&config)
	
	mux := routes.New()

	// attach routes to their respective handlers
	mux.Get("/profile/:email", GetProfile)
	mux.Post("/profile", PostProfile)
	mux.Put("/profile/:email", PutProfile)
	mux.Del("/profile/:email", DeleteProfile)

	// attach our routes to the root 
	http.Handle("/", mux)
	
	// start the rpc server
	log.Println("Starting RPC server")
	go ListenAndServeRPC(&config)
	
	// start the server
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", config.PortNum)
	log.Println("HTTP server listening at:", addr)
	err = http.ListenAndServe(addr, nil)
	if err!=nil{
		// error while starting the server
		fmt.Println(err.Error())	
	}
}