package main

//go:generate ./make_docs.sh
//go:generate go run readme_gen.go

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type HandlerWrapper struct {
	dh           DNSHandler
	lv           *LogViewer
	dottedDomain string
}

func (w *HandlerWrapper) ServeDNS(wr dns.ResponseWriter, r *dns.Msg) {
	msg := &dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true

	domain := strings.ToLower(msg.Question[0].Name)
	q := &query{
		t:            msg.Question[0].Qtype,
		name:         domain,
		nameForReply: domain,
	}

	replies, _ := w.dh.Handle(q)

	for _, s := range replies {
		tryAdd(msg, s)
	}

	if strings.HasSuffix(q.name, w.dottedDomain) && q.name != w.dottedDomain {
		lr := &LogRecord{
			Time:          time.Now(),
			RemoteAddr:    wr.RemoteAddr().String(),
			RequestType:   dns.TypeToString[q.t],
			RequestDomain: q.name,
			Replies:       replies,
		}

		w.lv.Push(lr)
	}

	if err := wr.WriteMsg(msg); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %v CONFIG", os.Args[0])
	}

	config, err := NewConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handlers := DNSHandlers{
		NewDelayRecordHandler(),
		NewNoHTTPSRecordHandler(),
		NewFakeRecordHandler(),
		NewPredefinedRecordHandler(config.PredefinedRecords),
		NewRebindForTimesRecordHandler(),
		NewRebindForRecordHandler(),
		NewRebindRecordHandler(),
		NewMakeRecordHandler(),
		NewIncRecordHandler(),
	}

	lv := NewLogViewer()
	if err := runHTTPServers(config, lv); err != nil {
		log.Fatal(err)
	}

	srv := &dns.Server{
		Addr: ":53",
		Net:  "udp",
		Handler: &HandlerWrapper{
			dh:           handlers,
			lv:           lv,
			dottedDomain: config.Domain + ".",
		},
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}

func runHTTPServers(config *Config, lv *LogViewer) error {
	mux := http.NewServeMux()
	lv.RegisterHandlers(mux)
	mux.Handle("/", http.FileServer(Readme))

	for _, addr := range config.HTTP.ListenOn {
		go func(addr string) {
			err := http.ListenAndServe(addr, mux)
			if err != nil {
				log.Fatal(err)
			}
		}(addr)
	}

	return nil
}
