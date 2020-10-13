package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"yall.in"
)

type Status struct {
	CameraOn bool      `json:"cameraOn"`
	LastSync time.Time `json:"lastSync"`
}

func getStatus(ctx context.Context) (*Status, error) {
	// the server keeps track of our state by our mac address
	// meaning we need to know our mac address
	macs, err := getMacAddr()
	if err != nil {
		return nil, fmt.Errorf("Error determining MAC address: %w", err)
	}
	if len(macs) < 1 {
		return nil, fmt.Errorf("can't find network interface with mac address")
	}
	request, err := http.NewRequest(http.MethodGet, "http://peter.local.carvers.house:9988/status/", nil)
	if err != nil {
		return nil, fmt.Errorf("Error building request: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving statuses from server: %w", err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		yall.FromContext(ctx).WithField("status", resp.Status).WithField("response", string(response)).Warn("unexpected response status")
		return nil, fmt.Errorf("Unexpected response: %s", resp.Status)
	}
	statuses := map[string]Status{}
	err = json.Unmarshal(response, &statuses)
	if err != nil {
		return nil, fmt.Errorf("Error parsing response: %w", err)
	}
	v, ok := statuses[macs[0]]
	if !ok {
		return nil, fmt.Errorf("Device hasn't reported status to the server.")
	}
	return &v, nil
}
