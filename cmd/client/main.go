package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Device struct {
	IP         string
	EndIP      int
	Port       []string
	PortOK     bool
	PingTimeMs int64
	Hostname   string
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
func (d *DevicePool) ping(ip string, endIP int, c chan int) {
	var ok bool
	var Ms int64
	var beforeTime int64
	var afterTime int64

	dev := Device{
		IP:         ip + strconv.Itoa(endIP),
		EndIP:      endIP,
		PingTimeMs: int64(d.OptionTimeOut),
	}

	if hostname, err := net.LookupAddr(dev.IP); err == nil {
		dev.Hostname = hostname[0]
	}

	for _, port := range d.OptionPorts {
		beforeTime = time.Now().UnixMilli()
		conn, err := net.DialTimeout("tcp", dev.IP+":"+port, time.Duration(d.OptionTimeOut)*time.Second)
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

func (d *DevicePool) flagParse(ipRanges *string, ports *string) {

	d.OptionIPBase = strings.Replace(*ipRanges, "x", "", 1)
	fmt.Println("Scan ip range:", d.OptionIPBase+"x")

	d.OptionPorts = strings.Split(*ports, ",")
	fmt.Println("Scan ports:", d.OptionPorts)
}

//
//
//
//
func (d *DevicePool) displayOK() {
	var deviceArray []Device

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"IP", "Ping", "Hostname", "Ports"})

	deviceArray = d.DeviceOK

	for i := 0; i < len(deviceArray)-1; i++ {
		for j := 0; j < len(deviceArray)-i-1; j++ {
			if deviceArray[j].EndIP > deviceArray[j+1].EndIP {
				deviceArray[j], deviceArray[j+1] = deviceArray[j+1], deviceArray[j]
			}
		}
	}

	for _, dev := range d.DeviceOK {
		t.AppendRow(table.Row{
			dev.IP,
			strconv.FormatInt(dev.PingTimeMs, 10) + " ms",
			dev.Hostname,
			dev.Port})
		t.AppendSeparator()
	}
	t.Render()
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
	var threads int
	var threadsDone int
	var ipRange = flag.String("ip", "192.168.1.x", "IP Range")
	var ports = flag.String("p", "80", "Ports to scan, example: 80,81")
	var timeout = flag.Int("t", 5, "Timeout in sec")
	//	var help = flag.Bool("help", false, "Display help page")

	flag.Parse()

	devPool := &DevicePool{
		OptionTimeOut: *timeout,
	}
	c := make(chan int)

	devPool.flagParse(ipRange, ports)

	for i := 1; i < 255; i++ {
		threads++
		go devPool.ping(devPool.OptionIPBase, i, c)
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
}
