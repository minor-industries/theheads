package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minor-industries/platform/common/discovery"
	"github.com/minor-industries/platform/common/util"
	"github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"regexp"
	"sort"
	"strings"
)

type journalSchema struct {
	Hostname         string `json:"_HOSTNAME"`
	Unit             string `json:"UNIT"`
	SyslogIdentifier string `json:"SYSLOG_IDENTIFIER"`
	SystemdUnit      string `json:"_SYSTEMD_UNIT"`
	Message          string `json:"MESSAGE"`
}

type StreamLogsCommand struct {
	X            []string `long:"x" description:"exclude matching regular expression"`
	Ne           []string `long:"ne" description:"exclude exactly matching string"`
	XUnit        []string `long:"xunit" description:"exclude unit"`
	XSyslogId    []string `long:"xsyslog-id" description:"exclude syslog id"`
	XSystemdUnit []string `long:"xsystemd-unit" description:"exclude syslog id"`
	Xhost        []string `long:"xhost" description:"exclude host"`
}

func (s *StreamLogsCommand) Execute(args []string) error {
	txt, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal")
	}

	fmt.Println(string(txt))

	var regexs []*regexp.Regexp

	for _, x := range s.X {
		r, err := regexp.Compile(x)
		if err != nil {
			return errors.Wrap(err, "compile regex")
		}
		regexs = append(regexs, r)
	}

	logger, _ := util.NewLogger(false)
	ch := make(chan string)

	services, err := discovery.NewSerf("127.0.0.1:7373").Discover(logger)
	if err != nil {
		return errors.Wrap(err, "discover")
	}

	for _, entry := range services {
		if entry.Service != "logstream" {
			continue
		}

		go func(entry *discovery.Entry) {
			addr := fmt.Sprintf("%s:%d", entry.Hostname, entry.Port)

			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				showLog(&journalSchema{
					Hostname: entry.Hostname,
					Message:  errors.Wrap(err, "ERROR DIALING HOST").Error(),
				}, nil)
				return
			}

			client := heads.NewLogstreamClient(conn)
			logs, err := client.StreamLogs(context.Background(), &heads.Empty{})
			if err != nil {
				showLog(&journalSchema{
					Hostname: entry.Hostname,
					Message:  errors.Wrap(err, "ERROR STREAMING LOGS").Error(),
				}, nil)
				return
			}

			for {
				msg, err := logs.Recv()
				if err != nil {
					showLog(&journalSchema{
						Hostname: entry.Hostname,
						Message:  errors.Wrap(err, "RECV ERROR").Error(),
					}, nil)
					return
				}

				ch <- msg.Log
			}
		}(entry)
	}

	for msg := range ch {
		log := &journalSchema{}
		err := json.Unmarshal([]byte(msg), log)
		if err != nil {
			fmt.Println("unmarshal json error")
			fmt.Println("")
			continue
		}

		zapLog := map[string]any{}
		_ = json.Unmarshal([]byte(log.Message), &zapLog)

		if filter(s, regexs, log, zapLog) {
			continue
		}

		showLog(log, zapLog)
	}

	return nil
}

func getMsg(
	log *journalSchema,
	zapLog map[string]any,
) string {
	m, ok := zapLog["msg"].(string)

	if ok {
		return m
	} else {
		return log.Message
	}
}

func filter(s *StreamLogsCommand, regexs []*regexp.Regexp, log *journalSchema, zapLog map[string]any) bool {
	msg := getMsg(log, zapLog)

	for _, ne := range s.Ne {
		if msg == ne {
			return true
		}
	}

	for _, x := range regexs {
		ok := x.MatchString(msg)
		if ok {
			return true
		}
	}

	for _, host := range s.Xhost {
		if host == log.Hostname {
			return true
		}
	}

	for _, unit := range s.XUnit {
		if log.Unit == unit {
			return true
		}
	}

	for _, id := range s.XSyslogId {
		if log.SyslogIdentifier == id {
			return true
		}
	}

	for _, id := range s.XSystemdUnit {
		if log.SystemdUnit == id {
			return true
		}
	}

	return false
}

func showLog(
	log *journalSchema,
	zapLog map[string]any,
) {
	msg, _ := zapLog["msg"].(string)

	if msg == "" {
		fmt.Printf(
			"host=%s unit=%s syslog-id=%s systemd-unit=%s\n%s\n\n",
			log.Hostname,
			log.Unit,
			log.SyslogIdentifier,
			log.SystemdUnit,
			"  "+strings.TrimSpace(log.Message),
		)
	} else {
		var show []string

		for k, v := range zapLog {
			if k == "msg" {
				continue
			}
			show = append(show, fmt.Sprintf("    %s: %v", k, v))
		}

		sort.Strings(show)

		show = append([]string{"  " + msg}, show...)

		fmt.Printf(
			"host=%s unit=%s syslog-id=%s systemd-unit=%s\n%s\n\n",
			log.Hostname,
			log.Unit,
			log.SyslogIdentifier,
			log.SystemdUnit,
			strings.Join(show, "\n"),
		)
	}
}
