package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"yall.in"
)

func update(ctx context.Context, on bool) error {
	// the server keeps track of our state by our mac address
	// meaning we need to know our mac address
	macs, err := getMacAddr()
	if err != nil {
		return fmt.Errorf("Error determining MAC address: %w", err)
	}
	if len(macs) < 1 {
		return fmt.Errorf("can't find network interface with mac address")
	}
	type req struct {
		CameraOn bool `json:"cameraOn"`
	}
	r := req{CameraOn: on}
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("Error building request body: %w", err)
	}
	buf := bytes.NewBuffer(b)
	log.Println("setting", macs[0], "to", on)
	// TODO: let's not hardcode the server in
	request, err := http.NewRequest(http.MethodPatch, "http://peter.local.carvers.house:9988/status/"+macs[0], buf)
	if err != nil {
		return fmt.Errorf("Error building request: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error updating server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response: %w", err)
	}
	yall.FromContext(ctx).WithField("status", resp.Status).WithField("response", string(response)).Warn("unexpected response status")
	return nil
}
