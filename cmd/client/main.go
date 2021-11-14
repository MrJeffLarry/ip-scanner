package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Device struct {
	IP         string
	Port       []string
	PortOK     bool
	PingTimeMs int64
}

type DevicePool struct {
	OptionDisplayFail     bool
	OptionDisplayRealTime bool
	OptionIPBase          string
	OptionTimeOut         int
	OptionPorts           []string
	DeviceOK              []Device
	DeviceFail            []Device
}

//
//
//
//
func (d *DevicePool) ping(ip string, c chan int) {
	var ok bool
	var Ms int64
	var beforeTime int64
	var afterTime int64

	dev := Device{
		IP:         ip,
		PingTimeMs: int64(d.OptionTimeOut),
	}

	for _, port := range d.OptionPorts {
		beforeTime = time.Now().UnixMilli()

		conn, err := net.DialTimeout("tcp", ip+":"+port, time.Duration(d.OptionTimeOut)*time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "connect: connection refused") {
				ok = true
			}
		} else {
			conn.Close()
			ok = true
			dev.Port = append(dev.Port, port)
		}
		afterTime = time.Now().UnixMilli()
		Ms = (afterTime - beforeTime)
		if Ms < dev.PingTimeMs {
			dev.PingTimeMs = Ms
		}
	}

	if ok {
		d.DeviceOK = append(d.DeviceOK, dev)
	} else {
		d.DeviceFail = append(d.DeviceFail, dev)
	}

	c <- 1
}

//
//
//
//
func (d *DevicePool) argParse() {
	args := os.Args[1:]
	if len(args) <= 1 {
		fmt.Println("No valid option: ipscan 192.168.1.x 80,81")
		os.Exit(2)
		return
	}

	d.OptionIPBase = strings.Replace(args[0], "x", "", 1)
	fmt.Println("Scan ip range:", d.OptionIPBase+"x")

	d.OptionPorts = strings.Split(args[1], ",")
	fmt.Println("Scan ports:", d.OptionPorts)

	hold := false

	for _, arg := range args {
		if hold {
			timeOut, err := strconv.Atoi(arg)
			if err == nil {
				fmt.Println("Scan ip timeout:", timeOut)
				d.OptionTimeOut = timeOut
			}
			hold = false
		} else {
			switch arg {
			case "dFail":
				d.OptionDisplayFail = true
			case "t":
				hold = true
			}
		}
	}
}

//
//
//
//
func (d *DevicePool) displayOK() {
	fmt.Println("************* Device found list **************")
	for _, dev := range d.DeviceOK {
		fmt.Println(dev.IP, dev.Port, dev.PingTimeMs, "ms")
	}
	fmt.Println("Found:", len(d.DeviceOK), "devices")
}

//
//
//
//
func (d *DevicePool) displayFail() {
	fmt.Println("************* Device not found list **************")
	for _, dev := range d.DeviceFail {
		fmt.Println(dev.IP, dev.Port)
	}
	fmt.Println("Not Found:", len(d.DeviceFail), "devices")
}

//
//
//
//
func main() {
	devPool := &DevicePool{
		OptionTimeOut: 5,
	}
	c := make(chan int)
	threads := 0
	threadsDone := 0

	devPool.argParse()

	for i := 1; i < 255; i++ {
		ip := strconv.Itoa(i)
		threads++
		go devPool.ping(devPool.OptionIPBase+ip, c)
	}

	// Wait for all rutines to complete
	for {
		threadsDone += <-c
		if threadsDone >= threads {
			break
		}
	}

	devPool.displayOK()

	if devPool.OptionDisplayFail {
		devPool.displayFail()
	}

	fmt.Println("************* Done **************")
}
