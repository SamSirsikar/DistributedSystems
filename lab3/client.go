package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

func hashValue(key string) uint32 {
    h := fnv.New32a()
	h.Write([]byte(key))
	return uint32(h.Sum32())
}

func hashNode(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func computeWeight(node uint32, key uint32) int {
    // based on the formula given in page 19
    // of the paper given in the link below:
    // http://www.eecs.umich.edu/techreports/cse/96/CSE-TR-316-96.pdf
    
    weight := (1103515245 * ((1103515245 * node + 12345) ^ key + 12345))
	if weight < 0 {
		weight += ((1 << 31) - 1)
	}
	return int(weight)
}

type Node struct {
	Key     string
	HashKey uint32
}

func NewNode(key string) *Node {
	return &Node{
		Key:     key,
		HashKey: hashNode(key),
	}
}

type Nodes []*Node

func (n Nodes) Len() int {
	return len(n)

}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
func (n Nodes) Less(i, j int) bool {
	return n[i].HashKey < n[j].HashKey
}

type HRWHashRing struct {
	Nodes Nodes
}

func NewHRWHashRing() *HRWHashRing {
	return &HRWHashRing{Nodes: Nodes{}}
}

func (c *HRWHashRing) Add(key string) {
	node := NewNode(key)
	c.Nodes = append(c.Nodes, node)
}

func (c *HRWHashRing) Remove(key string) error {
	i := c.search(key)
	if i >= c.Nodes.Len() || c.Nodes[i].Key != key {
		return errors.New("key not found")
	}

	c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)

	return nil
}

func (c *HRWHashRing) Get(key string) string {
    // 0. convert key to a number using murmur3 or fnv.
    // 1. iterate throught each node.
    // 2. compute the weight by passing key and node.
    // 3. find the max of all weights.
    // 4. return the node with max weight.
    
    hashKey := hashValue(key)
    maxWeight := 0
    maxIndex := 0
    for i := 0; i < len(c.Nodes); i++ {
        weight := computeWeight(c.Nodes[i].HashKey, hashKey)
        if weight > maxWeight {
            maxWeight = weight
            maxIndex = i
        }
    }
    
    return c.Nodes[maxIndex].Key
}

func (c *HRWHashRing) search(key string) int {
	return sort.Search(c.Nodes.Len(), func(i int) bool {
		return c.Nodes[i].HashKey >= hashNode(key)
	})
}

func doPut(url string) error {
	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, strings.NewReader(""))
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println("Response Status Code:", response.StatusCode)
	fmt.Println("Response Content:", contents)
	return nil
}

func main() {
	// generate port numbers
	if len(os.Args) < 3 {
		fmt.Println("{key}->{value} usage: go run client.go \"3001-3005\" \"1->A,2->B,3->C,4->D,5->E\"")
		// go run client.go "3001-3005" "1->A,2->B,3->C,4->D,5->E"
		os.Exit(1)
	}

	// get the start and end ports
	startEndPort := strings.Split(os.Args[1], "-")
	startPort, _ := strconv.Atoi(startEndPort[0])
	endPort, _ := strconv.Atoi(startEndPort[1])

	// create a HRW hash ring
	ch := NewHRWHashRing()

	for i := startPort; i <= endPort; i++ {
		// add each server to the consistent hash
		ch.Add(fmt.Sprintf("http://localhost:%d", i))
	}

	keyValuePairs := strings.Split(os.Args[2], ",")
	for i := 0; i < len(keyValuePairs); i++ {
		keyValue := strings.Split((keyValuePairs[i]), "->")

		// now, determine which server to send
		// based on the key
		url := ch.Get(keyValue[0])

		// now, make a request to this url using
		// the key and value as PUT url/key/val
		// example: PUT http://localhost:3001/1/A
		// will save the value A at key 1 on server 3001
		fmt.Printf("Sending %s to %s\n", keyValuePairs[i], url)
		err := doPut(fmt.Sprintf("%s/%s/%s", url, keyValue[0], keyValue[1]))
		if err != nil {
			fmt.Println("Request to", url, "failed")
		}
	}
}
