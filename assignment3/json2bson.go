package main

import (
    "gopkg.in/mgo.v2/bson"
    "encoding/json" 
    "log"
    "io/ioutil"
    "os"
)

type RequestData struct {
    Url string `json:"url" bson:"url"`
    Method string `json:"method" bson:"method"`
    HTTPHeaders map[string]string `json:"http_headers" bson:"http_headers"`
    Body map[string]string `json:"body" bson:"body"`
}

type InputData struct {
    Id int `json:"id" bson:"id"`
    SuccessHTTPResponseCode int `json:"success_http_response_code" bson:"success_http_response_code"`
    MaxRetries int `json:"max_retries" bson:"max_retries"`
    CallbackWebhookUrl string `json:"callback_webhook_url" bson:"callback_webhook_url"`
    Request RequestData `json:"request" bson:"request"`
}

func loadInput() (*InputData, error) {
    in, err := ioutil.ReadFile("./input.json")
    if err != nil {
        return nil, err
    }
    
    input := InputData{}
    err = json.Unmarshal(in, &input)
    if err != nil {
        return nil, err
    }
    return &input, nil
    
    
}

func dumpInput(input *InputData) error {
    file, err := os.Create("./input.bson")
    if err != nil {
        return err
    }
    // defer is called while exiting the function, 
    // even if there is an exception in the function
    defer file.Close()   

    data, err := bson.Marshal(input)
    if err != nil {
        return err
    }
    
    _, err = file.Write(data)
    if err != nil {
        return err
    }
    // writes from buffer to the file
    err = file.Sync()
    if err != nil {
        return err
    }
    
    return nil
}



func main() {
    input, err := loadInput()
    if err != nil {
        log.Fatal(err.Error())
    }
    err = dumpInput(input)
    if err != nil {
        log.Fatal(err.Error())
    }
}