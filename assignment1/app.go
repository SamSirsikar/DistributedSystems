package main

import (
	"github.com/drone/routes"
	"log"
	"net/http"
	"encoding/json"
	"bytes"
)


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
}


func (p *ProfileManager) Get(key string) *Profile {
	// reads data stored in the ProfileManager
	if profile, ok := p.Data[key]; ok {
  		return profile
	}
	return nil
}

func (p *ProfileManager) Set(key string, val *Profile) {
	// saves the data into the ProfileManager
	p.Data[key] = val
}

func (p *ProfileManager) UnSet(key string) {
	// removes the data from ProfileManager based on the key
	delete(p.Data, key)
}


func New() *ProfileManager {
	// creates and returns a new ProfileManager object
	pm := ProfileManager{}
	pm.Data = make(map[string]*Profile)
	
	return &pm
}

var profileManager *ProfileManager


func main() {
	// create a profile manager
	profileManager = New()
	
	mux := routes.New()

	// attach routes to their respective handlers
	mux.Get("/profile/:email", GetProfile)
	mux.Post("/profile", PostProfile)
	mux.Put("/profile/:email", PutProfile)
	mux.Del("/profile/:email", DeleteProfile)

	// attach our routes to the root 
	http.Handle("/", mux)
	
	log.Println("Listening...")
	//fmt.Printf("%s:%s", os.Getenv("IP"), os.Getenv("PORT"))
	http.ListenAndServe(":3000", nil)
	// start the server
	//http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("IP"), os.Getenv("PORT")), nil)
}

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
    profileManager.Set(profile.Email, &profile)
    
    // return 200 status
	w.WriteHeader(http.StatusCreated)
}

func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	// get email from the url
	params := r.URL.Query()
	email := params.Get(":email")
	
	// delete the user corresponding to the email
	profileManager.UnSet(email)
	
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
    profileManager.Set(profile.Email, profile)
	
	// return a 204
	w.WriteHeader(http.StatusNoContent)
}

