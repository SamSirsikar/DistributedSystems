Running Server

# the below command will start 
# 5 http servers in the port range specified
go run server.go 3001-3005
Testing the Server

Client

go run client.go "3001-3005" "1->A,2->B,3->C,4->D,5->E"

Testing

# get the data from server at 3003
bash get.sh 3003 1
bash get.sh 3003 2

# get all the data from server at 3003
bash get.sh 3003

