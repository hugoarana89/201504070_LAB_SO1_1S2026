package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const CARNET = "201504070"

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "UP",
		"message":   "API3 is Ready",
		"timestamp": time.Now().Format(time.RFC3339),
		"VM":        2,
		"carnet":    CARNET,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func callAPI1Handler(w http.ResponseWriter, r *http.Request) {
	callAPI(w, "API1", "http://192.168.122.101:8080/health")
}

func callAPI2Handler(w http.ResponseWriter, r *http.Request) {
	callAPI(w, "API2", "http://192.168.122.101:8080/health")
}

func callAPI(w http.ResponseWriter, apiName string, url string) {
	client := http.Client{Timeout: 5 * time.Second}
	response := map[string]interface{}{
		"apiname": apiName,
		"carnet":  CARNET,
	}

	resp, err := client.Get(url)
	if err != nil {
		response["message"] = fmt.Sprintf("ERROR: The %s is not working", apiName)
		response["connection"] = false
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	vmNumber := 0
	if vm, ok := data["VM"].(float64); ok {
		vmNumber = int(vm)
	}

	if data["status"] == "UP" {
		response["message"] = fmt.Sprintf("The %s located on the VM%d is working", apiName, vmNumber)
		response["connection"] = true
	} else {
		response["message"] = fmt.Sprintf("ERROR: The %s located on the VM%d is not working", apiName, vmNumber)
		response["connection"] = false
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc(fmt.Sprintf("/api3/%s/call-api1", CARNET), callAPI1Handler)
	http.HandleFunc(fmt.Sprintf("/api3/%s/call-api2", CARNET), callAPI2Handler)

	fmt.Println("API3 listening on :8080")
	http.ListenAndServe(":8080", nil)
}

