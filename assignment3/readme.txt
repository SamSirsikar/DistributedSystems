Running the script:

    go run app.go

Input file: input.bson
Output file: output.bson

How does it work:
    * The scheduler calls the task function every N seconds (as configured).
    * Each time the task is called, it performs the following operations in 
      the order specified here:
        1. Reads the output.bson file if it exists (if not, ignores).
        2. Processes the output.bson and checks if there is a job with
           status 'STILL_TRYING'. If it does, the job details is read and 
           a request is made to the URL specified until the job succeeds or
           the max retires reaches 0 (whichever is earlier).
        3. If the status is success upon request (ie. the job succeeds), then 
           the job is marked as 'COMPLETE'; if not it is marked as 'FAILED'.
        4. If there are no more pending jobs with 'STILL_TRYING' (ie. either the 
           output.bson doesn't exist or the job has status 'COMPLETE' or 
           'FAILED'), the input.bson file is read and a request is made to the
           URL specified in it, and the cycle repeats.

Additionally have written a code to convert json to bson: json2bson.go
