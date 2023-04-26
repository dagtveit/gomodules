package opareceiver

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	rcvr "go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type opaReceiver struct {
	cfg       *LogsConfig
	server    *http.Server
	logger    *zap.Logger
	consumer  consumer.Logs
	wg        *sync.WaitGroup
	id        component.ID // ID of the receiver component
	startTime time.Time
}

// newReceiver creates the Kubernetes events receiver with the given configuration.
func newLogsReceiver(params rcvr.CreateSettings, cfg *Config, consumer consumer.Logs) (*opaReceiver, error) {
	println("newLogsrec")
	recv := &opaReceiver{
		cfg:      &cfg.Logs,
		consumer: consumer,
		logger:   params.Logger,
		wg:       &sync.WaitGroup{},
		id:       params.ID,
	}
	println("loadconfig")
	//tlsConfig, err := recv.cfg.HTTP.TLSSetting.LoadTLSConfig()
	//if err != nil {
	//	return nil, err
	//}

	s := &http.Server{
		//TLSConfig:         tlsConfig,
		Handler:           http.HandlerFunc(recv.handleEvent),
		ReadHeaderTimeout: 20 * time.Second,
	}

	recv.server = s
	return recv, nil
}

func (l *opaReceiver) Start(ctx context.Context, host component.Host) error {
	return l.startListening(ctx, host)
}
func (ddr *opaReceiver) Shutdown(ctx context.Context) (err error) {
	return ddr.server.Shutdown(ctx)
}
func (l *opaReceiver) startListening(ctx context.Context, host component.Host) error {
	l.logger.Info("starting receiver HTTP server")
	// We use l.server.Serve* over l.server.ListenAndServe*
	// So that we can catch and return errors relating to binding to network interface on start.
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", l.cfg.HTTP.Endpoint)
	if err != nil {
		return err
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		l.logger.Info("Starting ServeHttp",
			zap.String("address", l.cfg.HTTP.Endpoint))
		//zap.String("certfile", l.cfg.HTTP.TLSSetting.CertFile),
		//zap.String("keyfile", l.cfg.HTTP.TLSSetting.KeyFile))
		println()
		println()
		println()
		err := l.server.Serve(listener)

		l.logger.Info("Serve HTTP done")

		if err != http.ErrServerClosed {
			l.logger.Error("ServeTLS failed", zap.Error(err))
			host.ReportFatalError(err)
		}
	}()
	return nil
}
func (opar *opaReceiver) handleEvent(rw http.ResponseWriter, req *http.Request) {
	var payload []byte
	if req.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			opar.logger.Info("Got payload with gzip, but failed to read", zap.Error(err))
			return
		}
		defer reader.Close()
		// Read the decompressed response body
		payload, err = io.ReadAll(reader)
		if err != nil {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			opar.logger.Info("Got payload with gzip, but failed to read", zap.Error(err))
			return
		}
	} else {
		var err error
		payload, err = io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			opar.logger.Info("Failed to read alerts payload", zap.Error(err), zap.String("remote", req.RemoteAddr))
			return
		}
	}

	opar.logger.Debug("Received decicion log from opa")
	logs, err := parsePayload(payload)
	if err != nil {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		opar.logger.Error("Failed to convert opa request payload to maps", zap.Error(err))
		return
	}

	if err := opar.consumer.ConsumeLogs(req.Context(), opar.processLogs(logs)); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		opar.logger.Error("Failed to consumer alert as log", zap.Error(err))
		return
	}
}
func (l *opaReceiver) processLogs(logs []opaDecPayload) plog.Logs {
	pLogs := plog.NewLogs()

	resourceLogs := pLogs.ResourceLogs().AppendEmpty()
	resource := resourceLogs.Resource()
	for _, log := range logs {
		resource.Attributes().PutStr("service.name", "dev-ingress") // log.Input.Attributes.Request.HTTP.Headers["apiname"]+"_logs"
		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		scopeLogs.Scope().SetName("otelcol/" + typeStr)

		logRecord := scopeLogs.LogRecords().AppendEmpty()
		logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(log.Timestamp))
		logRecord.Attributes().PutStr("email", log.Input.Attributes.Request.HTTP.Headers["email"])

		l.logger.Info(strconv.Itoa(len(log.Result.Meta)))
		if log.Result.Meta["traceparent"] != "" {
			l.logger.Debug("Setting transparent")
			logRecord.Attributes().PutStr("meta.annotation_type", "span_event")
			logRecord.Attributes().PutStr("meta.signal_type", "trace")

			var trace = strings.Split(log.Result.Meta["traceparent"], "-")
			logRecord.Attributes().PutStr("name", "ext_authz")
			logRecord.Attributes().PutStr("version", trace[0])
			logRecord.Attributes().PutStr("trace.trace_id", trace[1])
			logRecord.Attributes().PutStr("trace.parent_id", trace[2])
			logRecord.Attributes().PutStr("trace.trace_flags", trace[3])
			logRecord.Attributes().PutInt("severity_code", 9)
			logRecord.Attributes().PutStr("severity_text", "Information")
			logRecord.Attributes().PutStr("severity", "Information")

			l.logger.Debug(trace[0])
			l.logger.Debug(trace[1])
			l.logger.Debug(trace[2])
			l.logger.Debug(trace[3])
		} else {
			l.logger.Warn("unable to set traceparent")
		}

		j, _ := json.Marshal(log)
		//fmt.Println())
		logRecord.Body().SetStr(string(j))

	}

	return pLogs
}
func parsePayload(payload []byte) (jsonData []opaDecPayload, err error) {

	json.Unmarshal(payload, &jsonData)
	return jsonData, nil
}
