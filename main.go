package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type Config struct {
	SerialPort string `json:"serial_port"`
	HTTPPort   int    `json:"http_port"`
}

func loadConfig() Config {
	// Defaults
	cfg := Config{
		SerialPort: "/dev/mhi_DUN",
		HTTPPort:   5643,
	}
	// Try /etc/config/gosms-config.json first, then ./gosms-config.json
	paths := []string{"/etc/config/gosms-config.json", "gosms-config.json"}
	for _, path := range paths {
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()
			dec := json.NewDecoder(f)
			if err := dec.Decode(&cfg); err != nil {
				log.Printf("Error parsing config file %s, using defaults: %v", path, err)
			} else {
				log.Printf("Loaded config from %s", path)
				return cfg
			}
		}
	}
	log.Printf("Config file not found, using defaults")
	return cfg
}

func openSerialPort(path string) (io.ReadWriteCloser, error) {
	c := &serial.Config{
		Name:        path,
		Baud:        115200,
		ReadTimeout: time.Second * 5,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func readResponse(r io.Reader, timeout time.Duration) (string, error) {
	buf := make([]byte, 1024)
	var result strings.Builder

	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			break
		}
		n, err := r.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
			if strings.Contains(result.String(), "OK") || strings.Contains(result.String(), "ERROR") {
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return result.String(), nil
}

func sendATCommand(port io.ReadWriter, cmd string, timeout time.Duration) (string, error) {
	fullCmd := cmd + "\r"
	_, err := port.Write([]byte(fullCmd))
	if err != nil {
		return "", err
	}
	return readResponse(port, timeout)
}

func sendSMS(port io.ReadWriter, phone, message string) error {
	log.Printf("Sending SMS to %s: %s\n", phone, message)

	// Set SMS text mode
	resp, err := sendATCommand(port, "AT+CMGF=1", 3*time.Second)
	if err != nil {
		return fmt.Errorf("error setting text mode: %v", err)
	}
	log.Printf("Response: %s", resp)

	// Send SMS command
	cmd := fmt.Sprintf(`AT+CMGS="%s"`, phone)
	_, err = port.Write([]byte(cmd + "\r"))
	if err != nil {
		return err
	}

	// Send message + Ctrl+Z to send
	_, err = port.Write([]byte(message + string(rune(26)))) // 26 = Ctrl+Z
	if err != nil {
		return err
	}

	resp, err = readResponse(port, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error reading SMS send response: %v", err)
	}
	log.Printf("SMS send response: %s", resp)

	if strings.Contains(resp, "ERROR") {
		return fmt.Errorf("modem returned error sending SMS")
	}
	return nil
}

func main() {
	cfg := loadConfig()
	log.Printf("Using serial port: %s", cfg.SerialPort)
	log.Printf("Using HTTP port: %d", cfg.HTTPPort)

	port, err := openSerialPort(cfg.SerialPort)
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}
	defer port.Close()

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	http.HandleFunc("/send-sms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		phone := r.FormValue("phone")
		message := r.FormValue("message")

		if phone == "" || message == "" {
			http.Error(w, "Missing phone or message", http.StatusBadRequest)
			return
		}

		err := sendSMS(port, phone, message)
		if err != nil {
			log.Printf("Send SMS error: %v", err)
			http.Error(w, fmt.Sprintf("Failed to send SMS: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SMS sent successfully\n"))
	})

	addr := fmt.Sprintf("[::]:%d", cfg.HTTPPort)
	log.Printf("Listening on port %d (IPv4 and IPv6, if supported)...", cfg.HTTPPort)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
