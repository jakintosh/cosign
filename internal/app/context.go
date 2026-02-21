package app

import "net/http"

type RequestContext struct {
	IsHTMX      bool
	CurrentURL  string
	Target      string
	Trigger     string
	TriggerName string
	Boosted     bool
}

func requestContext(r *http.Request) RequestContext {
	return RequestContext{
		IsHTMX:      r.Header.Get("HX-Request") == "true",
		CurrentURL:  r.Header.Get("HX-Current-URL"),
		Target:      r.Header.Get("HX-Target"),
		Trigger:     r.Header.Get("HX-Trigger"),
		TriggerName: r.Header.Get("HX-Trigger-Name"),
		Boosted:     r.Header.Get("HX-Boosted") == "true",
	}
}
