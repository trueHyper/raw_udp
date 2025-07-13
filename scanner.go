package raw_udp

import (
	"context"
	"fmt"
	"log"
	//"net"
	"time"

	"github.com/zmap/zgrab2"
)

type Flags struct {
	zgrab2.BaseFlags

	SearchByPort   bool   `long:"use_port" description:"Enable serch proto by port"`
	DefaultPayload string `long:"default" description:"Name of payload to try first before the rest (e.g. --default default)"`
	Fast           bool   `long:"fast" description:"Enable fast search"`
	Set            string `long:"set" description:"Protocol set to scan (e.g. --set set1)"`
	Custom         string `long:"custom" description:"Specify custom list of protocol payloads to scan (e.g. --custom=\"snmpv1 ntp coap rdp\")"`
	//HardMode bool `long:"hard_mode" description:"Full bruteforce"`
	PayloadTimeout time.Duration `long:"payload-timeout" description:"Maximum time to wait for a single payload probe response" default:"1s"`
}

func (f *Flags) Help() string                 { return "" }
func (f *Flags) Validate(args []string) error { return nil }

type Module struct{}

func (m *Module) NewFlags() any              { return new(Flags) }
func (m *Module) NewScanner() zgrab2.Scanner { return new(Scanner) }
func (m *Module) Description() string        { return "Raw UDP Scanner" }

func init() {
	var module Module
	_, err := zgrab2.AddCommand("raw_udp", "Raw UDP Scanner", module.Description(), 132, &module)
	if err != nil {
		log.Fatal(err)
	}
	_build_sorted_ports_arr()
	_build_ports_map()
}

type Scanner struct {
	config            *Flags
	dialerGroupConfig *zgrab2.DialerGroupConfig
}

func (s *Scanner) Init(flags zgrab2.ScanFlags) error {
	f, _ := flags.(*Flags)
	s.config = f

	s.dialerGroupConfig = &zgrab2.DialerGroupConfig{
		TransportAgnosticDialerProtocol: zgrab2.TransportUDP,
		NeedSeparateL4Dialer:            false,
		BaseFlags:                       &f.BaseFlags,
	}

	return nil
}

func (s *Scanner) InitPerSender(senderID int) error { return nil }

func (s *Scanner) GetName() string    { return s.config.Name }
func (s *Scanner) GetTrigger() string { return s.config.Trigger }
func (s *Scanner) Protocol() string   { return "raw_udp" }
func (scanner *Scanner) GetDialerGroupConfig() *zgrab2.DialerGroupConfig {
	return scanner.dialerGroupConfig
}

func (scanner *Scanner) Scan(ctx context.Context, dialGroup *zgrab2.DialerGroup, t *zgrab2.ScanTarget) (zgrab2.ScanStatus, any, error) {
	sock, err := dialGroup.Dial(ctx, t)
	if err != nil {
		return zgrab2.TryGetScanStatus(err), nil, fmt.Errorf("could not connect to target %s: %w", t.String(), err)
	}
	defer zgrab2.CloseConnAndHandleError(sock)

	return GetRawUDPResponse(sock, t, scanner.config)
}
