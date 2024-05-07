package client

import (
	"fmt"
	"log/slog"
	"net/http/httptrace"
	"os"

	"github.com/rocket-pool/node-manager-core/log"
)

func createTracer(file *os.File, logger *slog.Logger) (*httptrace.ClientTrace, error) {
	tracer := &httptrace.ClientTrace{}
	tracer.ConnectDone = func(network, addr string, err error) {
		logger.Debug("HTTP Connect Done",
			slog.String("network", network),
			slog.String("addr", addr),
			log.Err(err),
		)
		err = writeToTraceFile(file, fmt.Sprintf("ConnectDone: network=%s, addr=%s, err=%v", network, addr, err))
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}
	tracer.DNSDone = func(dnsInfo httptrace.DNSDoneInfo) {
		logger.Debug("HTTP DNS Done",
			slog.String("addrs", fmt.Sprint(dnsInfo.Addrs)),
			slog.Bool("coalesced", dnsInfo.Coalesced),
			log.Err(dnsInfo.Err),
		)
		err := writeToTraceFile(file, fmt.Sprintf("DNSDone: addrs=%v, coalesced=%t, err=%v", dnsInfo.Addrs, dnsInfo.Coalesced, dnsInfo.Err))
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}
	tracer.DNSStart = func(dnsInfo httptrace.DNSStartInfo) {
		logger.Debug("HTTP DNS Start",
			slog.String("host", dnsInfo.Host),
		)
		err := writeToTraceFile(file, fmt.Sprintf("DNSStart: host=%s", dnsInfo.Host))
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}
	tracer.GotConn = func(connInfo httptrace.GotConnInfo) {
		logger.Debug("HTTP Got Connection",
			slog.Bool("reused", connInfo.Reused),
			slog.Bool("wasIdle", connInfo.WasIdle),
			slog.Duration("idleTime", connInfo.IdleTime),
			slog.String("localAddr", connInfo.Conn.LocalAddr().String()),
			slog.String("remoteAddr", connInfo.Conn.RemoteAddr().String()),
		)
		err := writeToTraceFile(file, fmt.Sprintf("GotConn: reused=%t, wasIdle=%t, idleTime=%s, localAddr=%s, remoteAddr=%s", connInfo.Reused, connInfo.WasIdle, connInfo.IdleTime, connInfo.Conn.LocalAddr().String(), connInfo.Conn.RemoteAddr().String()))
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}
	tracer.GotFirstResponseByte = func() {
		logger.Debug("HTTP Got First Response Byte")
		err := writeToTraceFile(file, "GotFirstResponseByte")
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}
	tracer.PutIdleConn = func(err error) {
		logger.Debug("HTTP Put Idle Connection",
			log.Err(err),
		)
		err = writeToTraceFile(file, fmt.Sprintf("PutIdleConn: err=%v", err))
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}
	tracer.WroteRequest = func(wroteInfo httptrace.WroteRequestInfo) {
		logger.Debug("HTTP Wrote Request",
			log.Err(wroteInfo.Err),
		)
		err := writeToTraceFile(file, fmt.Sprintf("WroteRequest: err=%v", wroteInfo.Err))
		if err != nil {
			logger.Debug("<error writing to HTTP trace file>", log.Err(err))
		}
	}

	return tracer, nil
}

func writeToTraceFile(file *os.File, data string) error {
	// Write the data
	_, err := file.WriteString(data + "\n")
	if err != nil {
		return fmt.Errorf("error writing to HTTP trace file [%s]: %w", file.Name(), err)
	}
	return nil
}
