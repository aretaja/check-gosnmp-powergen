// Copyright 2022 by Marko Punnar <marko[AT]aretaja.org>
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

// check-gosnmp-powergen is power generator status plugin for Icinga2 compatible systems

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aretaja/godevman"
	"github.com/aretaja/icingahelper"
	"github.com/kr/pretty"
)

// Version of release
const Version = "0.0.1"

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	// Parse cli arguments
	var host = flag.String("H", "", "<host ip>")
	var snmpVer = flag.Int("V", 2, "[snmp version] (1|2|3)")
	var snmpUser = flag.String("u", "public", "[username|community]")
	var snmpProt = flag.String("a", "MD5", "[authentication protocol] (NoAuth|MD5|SHA)5")
	var snmpPass = flag.String("A", "", "[authentication protocol pass phrase]")
	var snmpSlevel = flag.String("l", "authPriv", "[security level] (noAuthNoPriv|authNoPriv|authPriv)")
	var snmpPrivProt = flag.String("x", "DES", "[privacy protocol] (NoPriv|DES|AES|AES192|AES256|AES192C|AES256C)")
	var snmpPrivPass = flag.String("X", "", "[privacy protocol pass phrase]")
	var ctype = flag.String("t", "", "<check type>\n"+
		"\telectrical - check electrical parameters\n"+
		"\tengine - check engine parameters\n"+
		"\tcommon - check common status\n",
	)
	var wVolt = flag.String("wv", "215:245", "[warning level for mains and gen. voltage] (V). ctype - electrical")
	var cVolt = flag.String("cv", "210:250", "[critical level for mains and gen. voltage] (V). ctype - electrical")
	var wCur = flag.String("wc", "24", "[warning level for gen. current] (A). ctype - electrical")
	var cCur = flag.String("cc", "27", "[critical level for gen. current] (A). ctype - electrical")
	var wPow = flag.String("wp", "13", "[warning level for gen. power] (kW). ctype - electrical")
	var cPow = flag.String("cp", "15", "[critical level for gen. power] (kW). ctype - electrical")
	var wFreq = flag.String("wf", "48:52", "[warning level for gen. freq.] (Hz). ctype - electrical")
	var cFreq = flag.String("cf", "46:54", "[critical level for gen. freq.] (Hz). ctype - electrical")
	var wBat = flag.String("wb", "130:145", "[warning level for battery voltage] (V*10). ctype - engine")
	var cBat = flag.String("cb", "120:155", "[critical level for battery voltage] (V*10). ctype - engine")
	var wFuel = flag.String("wl", "20:100", "[warning level for fuel level] (%). ctype - engine")
	var cFuel = flag.String("cl", "10:100", "[critical level for fuel level] (%). ctype - engine")
	var wTemp = flag.String("wt", "98", "[warning level for coolant temp] (°C). ctype - engine")
	var cTemp = flag.String("ct", "104", "[critical level for coolant temp] (°C). ctype - engine")

	var dbg = flag.Bool("d", false, "Using this parameter will print out debug info")
	var ver = flag.Bool("v", false, "Using this parameter will display the version number and exit")

	flag.Parse()

	// Initialize new check object
	check := icingahelper.NewCheck("GEN")

	// Show version
	if *ver {
		fmt.Println("plugin version " + Version)
		os.Exit(check.RetVal())
	}

	// Exit if no host submitted
	if net.ParseIP(*host) == nil {
		log.Println("error: valid host ip is required")
		os.Exit(check.RetVal())
	}

	// Exit if no check type submitted
	if *ctype == "" {
		log.Println("error: check type required")
		os.Exit(check.RetVal())
	}

	params := godevman.Dparams{
		Ip: *host,
		SnmpCred: godevman.SnmpCred{
			Ver:      *snmpVer,
			User:     *snmpUser,
			Prot:     *snmpProt,
			Pass:     *snmpPass,
			Slevel:   *snmpSlevel,
			PrivProt: *snmpPrivProt,
			PrivPass: *snmpPrivPass,
		},
	}

	device, err := godevman.NewDevice(&params)
	if err != nil {
		log.Printf("error: godevman: %v", err)
		os.Exit(check.RetVal())
	}

	md := device.Morph()
	// DEBUG
	if *dbg {
		fmt.Printf("%# v\n", md)
	}

	d, ok := md.(godevman.DevGenReader)
	if !ok {
		log.Println("error: type is not godevman.DevGenReader")
		os.Exit(check.RetVal())
	}

	switch *ctype {
	case "common":
		res, err := getInfo(d, "Common", *dbg)
		if err != nil {
			log.Printf("error: %v", err)
			os.Exit(check.RetVal())
		}
		common(check, res, *dbg)
	case "electrical":
		res, err := getInfo(d, "Electrical", *dbg)
		if err != nil {
			os.Exit(check.RetVal())
		}
		err = electrical(check, res, *dbg, *wVolt, *cVolt, *wCur, *cCur, *wPow, *cPow, *wFreq, *cFreq)
		if err != nil {
			log.Printf("error: %v", err)
			os.Exit(check.RetVal())
		}
	case "engine":
		res, err := getInfo(d, "Engine", *dbg)
		if err != nil {
			os.Exit(check.RetVal())
		}
		err = engine(check, res, *dbg, *wBat, *cBat, *wFuel, *cFuel, *wTemp, *cTemp)
		if err != nil {
			log.Printf("error: %v", err)
			os.Exit(check.RetVal())
		}
	default:
		log.Printf("error: unknown check type - %s", *ctype)
		os.Exit(check.RetVal())
	}

	fmt.Print(check.FinalMsg())
	os.Exit(check.RetVal())
}

func getInfo(d godevman.DevGenReader, t string, dbg bool) (godevman.GenInfo, error) {
	res, err := d.GeneratorInfo([]string{t})
	if err != nil {
		return res, err
	}
	// DEBUG
	if dbg {
		fmt.Printf("%# v\n", pretty.Formatter(res))
	}
	return res, err
}

func common(c *icingahelper.IcingaCheck, i godevman.GenInfo, dbg bool) {
	c.SetRetVal(0)
	if i.GenMode.IsSet {
		val := i.GenMode.Value
		if val != "Auto" {
			c.SetRetVal(2)
		}
		c.AddMsg(c.RetVal(), fmt.Sprintf("Mode: %s", val), "")
	} else {
		level := c.RetVal()
		if level != 2 {
			level = 3
		}
		c.SetRetVal(level)
		c.AddMsg(level, "Mode: Na", "")
	}

	if i.BreakerState.IsSet {
		val := i.BreakerState.Value
		if val != "MainsOper" {
			c.SetRetVal(2)
		}
		c.AddMsg(c.RetVal(), fmt.Sprintf("Breaker: %s", val), "")
	} else {
		level := c.RetVal()
		if level != 2 {
			level = 3
		}
		c.SetRetVal(level)
		c.AddMsg(level, "Breaker: Na", "")
	}

	if i.EngineState.IsSet {
		val := i.EngineState.Value
		if val != "Ready" {
			c.SetRetVal(2)
		}
		c.AddMsg(c.RetVal(), fmt.Sprintf("Engine: %s", val), "")
	} else {
		level := c.RetVal()
		if level != 2 {
			level = 3
		}
		c.SetRetVal(level)
		c.AddMsg(level, "Engine: Na", "")
	}
}

func electrical(c *icingahelper.IcingaCheck, i godevman.GenInfo, dbg bool, wVolt, cVolt, wCur, cCur, wPow, cPow, wFreq, cFreq string) error {
	data := map[string]godevman.SensorVal{
		"Mains Voltage L1": i.MainsVoltL1,
		"Mains Voltage L2": i.MainsVoltL2,
		"Mains Voltage L3": i.MainsVoltL3,
		"Gen Voltage L1":   i.GenVoltL1,
		"Gen Voltage L2":   i.GenVoltL2,
		"Gen Voltage L3":   i.GenVoltL3,
		"Gen Current L1":   i.GenCurrentL1,
		"Gen Current L2":   i.GenCurrentL2,
		"Gen Current L3":   i.GenCurrentL3,
		"Gen Power":        i.GenPower,
		"Gen Frequency":    i.GenFreq,
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		switch data[k].Unit {
		case "V":
			if data[k].IsSet {
				val := data[k].Value
				level := 0
				if strings.Contains(k, "Mains") || val != 0 {
					l, err := c.AlarmLevel(int64(val), wVolt, cVolt)
					if err != nil {
						return fmt.Errorf("voltage alarm level error: %v", err)
					}
					level = l
				}

				c.AddMsg(level, fmt.Sprintf("%s: %dV", k, val), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", wVolt, wVolt, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		case "A":
			if data[k].IsSet {
				val := data[k].Value
				level, err := c.AlarmLevel(int64(val), wCur, cCur)
				if err != nil {
					return fmt.Errorf("current alarm level error: %v", err)
				}

				c.AddMsg(level, fmt.Sprintf("%s: %dA", k, val), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", wCur, cCur, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		case "kW":
			if data[k].IsSet {
				val := data[k].Value
				level, err := c.AlarmLevel(int64(val), wPow, cPow)
				if err != nil {
					return fmt.Errorf("power alarm level error: %v", err)
				}

				c.AddMsg(level, fmt.Sprintf("%s: %dkW", k, val), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", wPow, cPow, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		case "Hz":
			if data[k].IsSet {
				val := data[k].Value
				level := 0
				if val != 0 {
					l, err := c.AlarmLevel(int64(val), wFreq, cFreq)
					if err != nil {
						return fmt.Errorf("power alarm level error: %v", err)
					}
					level = l
				}

				rVal := float64(val) / float64(data[k].Divisor)
				c.AddMsg(level, fmt.Sprintf("%s: %.1fHz", k, rVal), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", wFreq, cFreq, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		default:
			return fmt.Errorf("unexpected results from godevman")
		}
	}

	return nil
}

func engine(c *icingahelper.IcingaCheck, i godevman.GenInfo, dbg bool, wBat, cBat, wFuel, cFuel, wTemp, cTemp string) error {
	data := map[string]godevman.SensorVal{
		"Running Hours":       i.RunHours,
		"Fuel level":          i.FuelLevel,
		"Fuel Consumption":    i.FuelConsum,
		"Battery Voltage":     i.BatteryVolt,
		"Coolant Temperature": i.CoolantTemp,
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		switch k {
		case "Battery Voltage":
			if data[k].IsSet {
				val := data[k].Value
				level, err := c.AlarmLevel(int64(val), wBat, cBat)
				if err != nil {
					return fmt.Errorf("voltage alarm level error: %v", err)
				}

				rVal := float64(val) / float64(data[k].Divisor)
				c.AddMsg(level, fmt.Sprintf("%s: %.1fV", k, rVal), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", wBat, cBat, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		case "Coolant Temperature":
			if data[k].IsSet {
				val := data[k].Value
				level, err := c.AlarmLevel(int64(val), wTemp, cTemp)
				if err != nil {
					return fmt.Errorf("coolant alarm level error: %v", err)
				}

				c.AddMsg(level, fmt.Sprintf("%s: %d%s", k, val, data[k].Unit), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", wTemp, cTemp, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		case "Fuel level":
			if data[k].IsSet {
				val := data[k].Value
				level, err := c.AlarmLevel(int64(val), wFuel, cFuel)
				if err != nil {
					return fmt.Errorf("fuel alarm level error: %v", err)
				}

				c.AddMsg(level, fmt.Sprintf("%s: %d%s", k, val, data[k].Unit), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), data[k].Unit, wFuel, cFuel, "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		default:
			if data[k].IsSet {
				val := data[k].Value

				rVal := float64(val)
				if data[k].Divisor != 0 {
					rVal = rVal / float64(data[k].Divisor)
				}

				c.AddMsg(0, fmt.Sprintf("%s: %.1f%s", k, rVal, data[k].Unit), "")
				c.AddPerfData(fmt.Sprintf("'%s'", k), strconv.Itoa(int(val)), "", "", "", "0", "")
			} else {
				c.AddMsg(3, fmt.Sprintf("%s: Na", k), "")
			}
		}
	}

	name := "Number of Starts"
	if i.NumStarts.IsSet {
		val := i.NumStarts.Value
		c.AddMsg(0, fmt.Sprintf("%s: %d", name, val), "")
		c.AddPerfData(fmt.Sprintf("'%s'", name), strconv.Itoa(int(val)), "", "", "", "0", "")
	} else {
		level := c.RetVal()
		if level != 2 {
			level = 3
		}
		c.AddMsg(level, fmt.Sprintf("%s: Na", name), "")
	}

	return nil
}
