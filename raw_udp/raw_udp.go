package raw_udp

import (
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zmap/zgrab2"
)

const (
	BUFFER_SIZE   = 512
	EMPHTY_STR    = ""
	BAD_PROTO     = "bad_proto"
	UNKNOWN_PROTO = "unknown_proto"
)

type Result struct {
	Response string `json:"response,omitempty"`
	Proto    string `json:"proto,omitempty"`
}

func GetRawUDPResponse(conn net.Conn, t *zgrab2.ScanTarget, config *Flags) (zgrab2.ScanStatus, any, error) {
	var (
		response []byte
		proto    string
		err      error
	)

	if config.DefaultPayload != EMPHTY_STR {
		response, err = SendDefaultPayload(conn, config, int(t.Port))
		if err == nil && len(response) > 0 {
			return zgrab2.SCAN_SUCCESS, &Result{Response: string(response), Proto: config.DefaultPayload}, nil
		}
	}

	switch {
	case config.Custom != EMPHTY_STR:
		response, proto, err = SendCustomPayload(conn, config, int(t.Port))

	case config.Set != EMPHTY_STR:
		log.Infof("\033[32m•SEARCH BY SET(%s) ACTIVATE\033[0m\n", config.Set)
		response, proto, err = SearchBySet(conn, int(t.Port), config)

	case config.SearchByPort:
		log.Info("\033[32m•SEARCH BY PORT ACTIVATE\033[0m")
		response, proto, err = SearchByPort(conn, int(t.Port), config)
	default:
		// default action
	}

	if err != nil {
		fmt.Println(err)
		return zgrab2.SCAN_UNKNOWN_ERROR, nil, err
	}

	//encoded := base64.StdEncoding.EncodeToString(response)
	//log.Info("Base64:", encoded)
	//log.Info("Raw:", string(response))

	return zgrab2.SCAN_SUCCESS, &Result{Response: string(response), Proto: proto}, nil
}

func tryPayload(conn net.Conn, payload []byte, timeout time.Duration, proto string) ([]byte, error) {
	_, err := conn.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to write: %w", err)
	}

	if timeout > 0 {
		err = conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	buffer := make([]byte, BUFFER_SIZE)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	return buffer[:n], nil
}

func tryProtosFast(conn net.Conn, protos []string, timeout time.Duration) ([]byte, error) {
	for _, proto := range protos {
		payload, ok := _payload_[proto]
		if !ok {
			continue
		}
		_, err := conn.Write(payload.Data)
		if err != nil {
			log.Warnf("failed to write payload for %s: %v", proto, err)
		}
	}

	buf := make([]byte, BUFFER_SIZE)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("no response in fast mode: %w", err)
	}
	resp := buf[:n]
	log.Infof("\033[33m•RECEIVED RESPONSE IN FAST MODE (\033[35mLEN=%d\033[0m)\033[0m", len(resp))
	return resp, nil
}

func tryProtosSlow(conn net.Conn, protos []string, config *Flags, port int) ([]byte, string, error) {
	var bestResp []byte
	var prot string
	for _, proto := range protos {
		payload, ok := _payload_[proto]
		if !ok {
			log.Warnf("payload for proto %s not found", proto)
			continue
		}
		resp, err := tryPayload(conn, payload.Data, config.PayloadTimeout, proto)
		log.Infof("proto: %s, response: (\033[35mLEN=%d\033[0m) %v", proto, len(resp), resp)
		if err == nil && len(resp) >= len(bestResp) {
			bestResp = resp
			prot = proto
		}
	}
	if len(bestResp) > 0 {
		log.Infof("\033[33m•SELECTED PROTOCOL: %s\033[0m\n", prot)
		return bestResp, prot, nil
	}
	return nil, BAD_PROTO, fmt.Errorf("no valid response for port %d", port)
}

func SearchByPort(conn net.Conn, port int, config *Flags) ([]byte, string, error) {
	if protos, ok := LoadedPorts[port]; ok {
		log.Infof("PROTOS BY PORT FOUND: %d - %v\n", port, protos)

		if config.Fast {
			resp, err := tryProtosFast(conn, protos, config.PayloadTimeout)
			return resp, UNKNOWN_PROTO, err
		} else {
			return tryProtosSlow(conn, protos, config, port)
		}
	}

	if nearbyPort, err := GetNearbyPort(port); err == nil {
		log.Info("\033[33m•NERBY PORT SEARCHING...\033[0m")

		if config.Fast {
			resp, err := tryProtosFast(conn, LoadedPorts[nearbyPort], config.PayloadTimeout)
			return resp, UNKNOWN_PROTO, err
		} else {
			return tryProtosSlow(conn, LoadedPorts[nearbyPort], config, port)
		}
	}

	return nil, BAD_PROTO, fmt.Errorf("no payload found for port %d or nearby", port)
}

func SearchBySet(conn net.Conn, port int, config *Flags) ([]byte, string, error) {
	protos, ok := Sets[config.Set]
	if !ok {
		return nil, BAD_PROTO, fmt.Errorf("set %d not found", config.Set)
	}

	if config.Fast {
		resp, err := tryProtosFast(conn, protos, config.PayloadTimeout)
		return resp, UNKNOWN_PROTO, err
	} else {
		return tryProtosSlow(conn, protos, config, port)
	}
}

func SendDefaultPayload(conn net.Conn, config *Flags, port int) ([]byte, error) {
	if _, ok := _payload_[config.DefaultPayload]; ok {
		log.Infof("\033[32m•DEFAULT PAYLOAD IS SELECTED \033[35m%v\033[0m", config.DefaultPayload)
		resp, _, err := tryProtosSlow(conn, []string{config.DefaultPayload}, config, port)
		if err != nil {
			log.Warn(err)
		}
		return resp, err
	}
	log.Warnf("\033[31m•DEFAULT PAYLOAD \033[35m%v\033[31m IS NOT EXIST\033[0m", config.DefaultPayload)
	return nil, fmt.Errorf("default payload is not exist")
}

func SendCustomPayload(conn net.Conn, config *Flags, port int) ([]byte, string, error) {
	log.Infof("\033[32m•SEARCH BY CUSTOM LIST \033[35m%v\033[0m", config.Custom)

	protos := strings.Split(config.Custom, " ")
	Sets["custom_set"] = protos
	config.Set = "custom_set"

	return SearchBySet(conn, port, config)
}
