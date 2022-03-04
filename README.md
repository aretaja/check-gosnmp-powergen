# check-gosnmp-powergen
Icinga2 plugin designed to check power generator state.

Currently has support for ComAp InteliLite-14 generators (sysObjectId: 1.3.6.1.4.1.28634.14).
Uses experimental godevman module which is not yet available publicly.

Usage:
```
Usage of check-gosnmp-powergen:
  -A string
        [authentication protocol pass phrase]
  -H string
        <host ip>
  -V int
        [snmp version] (1|2|3) (default 2)
  -X string
        [privacy protocol pass phrase]
  -a string
        [authentication protocol] (NoAuth|MD5|SHA)5 (default "MD5")
  -cb string
        [critical level for battery voltage] (V*10). ctype - engine (default "120:155")
  -cc string
        [critical level for gen. current] (A). ctype - electrical (default "27")
  -cf string
        [critical level for gen. freq.] (Hz). ctype - electrical (default "46:54")
  -cl string
        [critical level for fuel level] (%). ctype - engine (default "10:100")
  -cp string
        [critical level for gen. power] (kW). ctype - electrical (default "15")
  -ct string
        [critical level for coolant temp] (°C). ctype - engine (default "104")
  -cv string
        [critical level for mains and gen. voltage] (V). ctype - electrical (default "210:250")
  -d    Using this parameter will print out debug info
  -l string
        [security level] (noAuthNoPriv|authNoPriv|authPriv) (default "authPriv")
  -t string
        <check type>
                electrical - check electrical parameters
                engine - check engine parameters
                common - check common status
    
  -u string
        [username|community] (default "public")
  -v    Using this parameter will display the version number and exit
  -wb string
        [warning level for battery voltage] (V*10). ctype - engine (default "130:145")
  -wc string
        [warning level for gen. current] (A). ctype - electrical (default "24")
  -wf string
        [warning level for gen. freq.] (Hz). ctype - electrical (default "48:52")
  -wl string
        [warning level for fuel level] (%). ctype - engine (default "20:100")
  -wp string
        [warning level for gen. power] (kW). ctype - electrical (default "13")
  -wt string
        [warning level for coolant temp] (°C). ctype - engine (default "98")
  -wv string
        [warning level for mains and gen. voltage] (V). ctype - electrical (default "215:245")
  -x string
        [privacy protocol] (NoPriv|DES|AES|AES192|AES256|AES192C|AES256C) (default "DES")
```

Examples:
```
$ ./check-gosnmp-powergen  -H 192.168.1.2 -u public -t common
GEN: OK - Mode: Auto; Breaker: MainsOper; Engine: Ready
```
```
$ ./check-gosnmp-powergen  -H 192.168.1.2 -u public -t electrical
GEN: OK - Gen Current L1: 0A; Gen Current L2: 0A; Gen Current L3: 0A; Gen Frequency: 0.0Hz; Gen Power: 0kW; Gen Voltage L1: 0V; Gen Voltage L2: 0V; Gen Voltage L3: 0V; Mains Voltage L1: 231V; Mains Voltage L2: 230V; Mains Voltage L3: 227V |'Gen Current L1'=0;24;27;0; 'Gen Current L2'=0;24;27;0; 'Gen Current L3'=0;24;27;0; 'Gen Frequency'=0;48:52;46:54;0; 'Gen Power'=0;13;15;0; 'Gen Voltage L1'=0;215:245;215:245;0; 'Gen Voltage L2'=0;215:245;215:245;0; 'Gen Voltage L3'=0;215:245;215:245;0; 'Mains Voltage L1'=231;215:245;215:245;0; 'Mains Voltage L2'=230;215:245;215:245;0; 'Mains Voltage L3'=227;215:245;215:245;0;
```
```
$ ./check-gosnmp-powergen  -H 10.55.104.117 -u HgsvQ92z -t engine
GEN: OK - Battery Voltage: 13.9V; Coolant Temperature: 59°C; Fuel Consumption: 0.0l; Fuel level: 100%; Running Hours: 0.2h; Number of Starts: 4 |'Battery Voltage'=139;130:145;120:155;0; 'Coolant Temperature'=59;98;104;0; 'Fuel Consumption'=0;;;0; 'Fuel level'=100%;20:100;10:100;0; 'Running Hours'=2;;;0; 'Number of Starts'=4;;;0;
```
