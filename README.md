# GoSMS Gateway

GoSMS Gateway is a lightweight, cross-platform SMS gateway written in Go, designed for sending SMS messages via SIM modules (cellular modems). It provides a simple HTTP API for SMS on OpenWRT routers and other Linux, Windows, or embedded devices.

*This project has been tested on OpenWRT and the GL-X3000 router, but should work on any device with a compatible SIM module.*

## Features
- Simple HTTP API for sending SMS via SIM modules (cellular modems)
- Optimized for OpenWRT routers
- Fast and efficient Go binary
- Easy to deploy on Linux, Windows, or embedded devices
- No external dependencies required for most use cases



## Download Pre-built Binaries
Pre-built and UPX-compressed binaries for all major platforms are available at:

https://github.com/HexiDev/GoSMS-Gateway/releases

---

## Installation (OpenWRT)
Follow these steps to install GoSMS Gateway on your OpenWRT device:

1. **Log in to your OpenWRT device**
2. **Copy the GoSMS Gateway binary to your device**
   - Place the compiled `send-sms-arm64` (or your architecture's binary) somewhere persistent, e.g. `/etc/config` or `/usr/bin`:
     ```sh
     scp send-sms-arm64 root@<router-ip>:/etc/config/
     ```
3. **Make the binary executable:**
   ```sh
   chmod +x /etc/config/send-sms-arm64
   ```
4. **Place the init script on your device**
   - Copy the init script from this repo's `init.d` directory to `/etc/init.d` on your device:
     ```sh
     scp init.d/sms-go root@<router-ip>:/etc/init.d/
     chmod +x /etc/init.d/sms-go
     ```
5. **Enable and start the service:**
   ```sh
   service sms-go enable
   service sms-go start
   # or reboot
   ```
6. **(Optional) Add or edit configuration:**
   - To customize the serial port or HTTP port, create a file named `gosms-config.json` and place it in `/etc/config` (or the same directory as the binary).
   - Example:
     ```sh
     nano /etc/config/gosms-config.json
     ```
   - See the Configuration section below for the file format and options. An example `gosms-config.json` is included in this repository.

---


## Usage
1. Download or copy the GoSMS Gateway binary for your platform (e.g., `send-sms-arm64` for OpenWRT, or another from the releases page).
2. Start the GoSMS Gateway service on your device:
   - On OpenWRT (recommended):
     ```sh
     service sms-go start
     ```
   - Or, for testing purposes, you can run the binary directly:
     ```sh
     ./send-sms-arm64
     ```
3. Send an SMS via HTTP POST to the running gateway:
   - Endpoint: `http://<host>:5643/send-sms`
   - Use form fields `phone` and `message`.
   - Example using `curl`:
     ```sh
     curl -X POST \
       -d "phone=+1234567890" \
       -d "message=Hello from GoSMS Gateway!" \
       http://localhost:5643/send-sms
     ```

If successful, the response will be:

```
SMS sent successfully
```

---


## Configuration
GoSMS Gateway can be configured using a `gosms-config.json` file. Place this file in `/etc/config` (recommended for OpenWRT) or in the same directory as the binary you are running. If the file is not present, defaults will be used.

Example `gosms-config.json`:
```json
{
  "serial_port": "/dev/mhi_DUN",
  "http_port": 5643
}
```

- `serial_port`: Path to the serial port device (e.g., `/dev/mhi_DUN` on Linux or `COM3` on Windows). Set this to match your modem's device path.
- `http_port`: Port for the HTTP server (default: 5643). Change this if you want the API to listen on a different port.

**How to use:**
- Copy the example above to a file named `gosms-config.json`.
- Edit the values as needed for your hardware and network.
- Place `gosms-config.json` in `/etc/config` (recommended) or in the same directory as the GoSMS Gateway binary before starting the service or running the binary.

If a value is not specified, the default will be used.

---

## How It Works
GoSMS Gateway communicates with a cellular modem using AT commands over a serial port. When an HTTP POST request is received on `/send-sms`, the gateway:

1. Opens the configured serial port to the modem.
2. Sends a sequence of AT commands to set up and send the SMS message.
3. Waits for the modem's response and returns the result to the HTTP client.

## AT Commands Used
- `AT+CMGF=1` — Set SMS mode to text.
- `AT+CMGS="<phone>"` — Initiate sending an SMS to the specified phone number.
- `<message><Ctrl+Z>` — Send the message text followed by Ctrl+Z (ASCII 26) to submit the SMS.

The gateway waits for `OK` or `ERROR` responses from the modem after each command to ensure reliable operation.

---

## Build (Optional)
To build and compress the GoSMS Gateway binary for OpenWRT (or other platforms):

1. Build the binary with Go (optimized for size):
   - For OpenWRT ARM64, use the following command to cross-compile:
     ```sh
     GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o send-sms-arm64 main.go
     ```
   - For your current platform:
     ```sh
     go build -ldflags="-s -w" -o send-sms-arm64 main.go
     ```
2. (Optional) Compress the binary with UPX for even smaller size:
   ```sh
   .\upx.exe --best send-sms-arm64
   ```

---

## License
MIT License

---
GoSMS Gateway is designed for reliability and simplicity. For more details, see the source code or contact the maintainer.
