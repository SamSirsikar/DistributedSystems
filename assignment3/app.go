package main

import (
    "fmt"
    "os"
    "time"
    "log"
    "bytes"
    "io/ioutil"
    "net/http"
    "gopkg.in/mgo.v2/bson"
    //"encoding/json"
    "github.com/jasonlvhit/gocron"
)

type RequestData struct {
    Url string `bson:"url"`
    Method string `bson:"method"`
    HTTPHeaders map[string]string `bson:"http_headers"`
    Body map[string]string `bson:"body"`
}

type InputData struct {
    Id int `bson:"id"`
    SuccessHTTPResponseCode int `bson:"success_http_response_code"`
    MaxRetries int `bson:"max_retries"`
    CallbackWebhookUrl string `bson:"callback_webhook_url"`
    Request RequestData `bson:"request"`
}

type ResponseData struct {
    HTTPResponseCode int `bson:"http_response_code"`
    HTTPHeaders map[string]string `bson:"http_headers"`
    Body map[string]string `bson:"body"`
}

type OutputData struct {
	Job struct {
		Status string `bson:"status"`
		NumRetries int `bson:"num_retries"`
	} `bson:"job"`
	Input InputData `bson:"input"`
    Output struct {
        Response ResponseData `bson:"response"`
    } `bson:"output"`
    CallbackResponseCode int `bson:"callback_response_code"`
}

func getCurrentTime() string {
    return time.Now().String()
}

func loadInput() (*InputData, error) {
    in, err := ioutil.ReadFile("./input.bson")
    if err != nil {
        return nil, err
    }
    
    input := InputData{}
    err = bson.Unmarshal(in, &input)
    if err != nil {
        return nil, err
    }
   
    return &input, nil
    
    
}

func loadOutput() (*OutputData, error) {
    out, err := ioutil.ReadFile("./output.bson")
    if err != nil {
        return nil, err
    }
    
    output := OutputData{}
    err = bson.Unmarshal(out, &output)
    if err != nil {
        return nil, err
    }
    
    return &output, nil
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

func dumpOutput(output *OutputData) error {
    file, err := os.Create("./output.bson")
    if err != nil {
        return err
    }
    // defer is called while exiting the function, 
    // even if there is an exception in the function
    defer file.Close()   

    log.Println(output)
    data, err := bson.Marshal(output)
    if err != nil {
        return err
    }
    
    _, err = file.Write(data)
    if err != nil {
        return err
    }
    err = file.Sync()
    if err != nil {
        return err
    }
    
    return nil
}

func makeRequest(request *RequestData) (*ResponseData, error) {
    // set the request payload
    data, err := bson.Marshal(request.Body)
    if err != nil {
        return nil, err
    }
    payload := bytes.NewBuffer(data)
    
    // configure a request
    req, err := http.NewRequest(request.Method, request.Url, payload)
    if err != nil {
        return nil, err
    }
    
    // set the request headers
    for key, val := range request.HTTPHeaders {
        req.Header.Set(key, val)
    }
    
    // create  a request client
    client := &http.Client{}
    
    // make the request
    res, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    response := ResponseData{}
    response.HTTPResponseCode = res.StatusCode
    response.HTTPHeaders = make(map[string]string)
    for key, val := range res.Header {
        response.HTTPHeaders[key] = val[0]
    }
    
    // load response body to a hashmap
    response.Body = make(map[string]string)
    buf := bytes.Buffer{}
    buf.ReadFrom(res.Body)
    err = bson.Unmarshal(buf.Bytes(), &response.Body)
    if err != nil {
        // assign a default json
        bson.Unmarshal([]byte(`{}`), &response.Body)
    }
    
    return &response, nil
}

func task() {
    // log the time at which this task was run
    t := getCurrentTime()
    fmt.Println("Running task @", t)
    
    // load the output file
    output, err := loadOutput()
    if err != nil {
        // print a log message and continue with input file
        log.Println(err)
    } else {
        // process the output file
        if output.Job.Status == "STILL_TRYING" {
            // make the request
            response, err := makeRequest(&output.Input.Request)
            if err != nil {
                log.Fatal(err)
            }
            
            // write the response to output.bson

            output.Output.Response = *response
            
            if response.HTTPResponseCode != output.Input.SuccessHTTPResponseCode {
               
                output.Job.NumRetries = output.Job.NumRetries - 1
                if output.Job.NumRetries == 0 {
                    output.Job.Status = "FAILED"   
                }
            } else {
                output.Job.NumRetries = 0
                output.Job.Status = "COMPLETED"
                
                // since status is 200, make a request to callback
                response, err = makeRequest(&RequestData{
                    Url: output.Input.CallbackWebhookUrl,
                    Method: "POST",
                })
                if err != nil {
                    log.Fatal(err)
                }
                output.CallbackResponseCode = response.HTTPResponseCode
            }
        
            err = dumpOutput(output)
            if err != nil {
                log.Fatal(err)
            }
            return
        }
        // NOTE: nothing to do when status is 'COMPLETED' or 'FAILED'
    }
    // load the input file
    input, err := loadInput()
    if err != nil {
        log.Fatal(err)
    }
    
    // make a request
    response, err := makeRequest(&input.Request)
    if err != nil {
        log.Fatal(err)
    }
    
    // write the response to output.bson
    output = &OutputData{}
    
    output.Input = *input
    output.Output.Response = *response

    if response.HTTPResponseCode != output.Input.SuccessHTTPResponseCode {
        output.Job.NumRetries = input.MaxRetries
        output.Job.Status = "STILL_TRYING"
    } else {
        output.Job.NumRetries = 0
        output.Job.Status = "COMPLETED"
        
        // since status is 200, make a request to callback
        response, err = makeRequest(&RequestData{
            Url: input.CallbackWebhookUrl,
            Method: "POST",
        })
        if err != nil {
            log.Fatal(err)
        }
        output.CallbackResponseCode = response.HTTPResponseCode
    }

    err = dumpOutput(output)
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    s := gocron.NewScheduler()
    s.Every(3).Seconds().Do(task)
    <- s.Start()
}

