package main

import (
	"context"
	"fmt"
	"sync"
)

type Asset struct{ ID, Title string }
type Job struct{ ID, AssetID, State string }
type Delivery struct{ Creator, Phone, MessageID string }

type Pipeline struct {
	SMS    *SMSClient
	mu     sync.Mutex
	assets map[string]Asset
	jobs   map[string]Job
}

func NewPipeline(sms *SMSClient) *Pipeline {
	return &Pipeline{SMS: sms, assets: map[string]Asset{}, jobs: map[string]Job{}}
}

func (p *Pipeline) IngestAndNotify(ctx context.Context, asset Asset, creator, phone string) (Job, Delivery, error) {
	p.mu.Lock()
	p.assets[asset.ID] = asset
	job := Job{ID: "job-" + asset.ID, AssetID: asset.ID, State: "processed"}
	p.jobs[job.ID] = job
	p.mu.Unlock()
	message := fmt.Sprintf("Asset %s is ready", asset.Title)
	id, err := p.SMS.Send(ctx, phone, message)
	if err != nil {
		return Job{}, Delivery{}, err
	}
	return job, Delivery{Creator: creator, Phone: phone, MessageID: id}, nil
}
