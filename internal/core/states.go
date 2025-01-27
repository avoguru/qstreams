package core

type StreamState string

const (
	Submitted StreamState = "submitted"
	Creating  StreamState = "creating"
	Running   StreamState = "running"
	Stopped   StreamState = "stopped"
)