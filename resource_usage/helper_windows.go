package resource_usage

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	wmic_command "github.com/go-zero-boilerplate/wmic-xml-command"
	"github.com/gocarina/gocsv"
)

type winHelper struct{}

func (w *winHelper) extractSingleValue(propToExtract string, wmicArgs []string) (string, error) {
	responseXML, err := wmic_command.Run(wmicArgs)
	if err != nil {
		return "", err
	}

	propValue := ""
	for _, res := range responseXML.Results {
		for _, prop := range res.Properties {
			if strings.EqualFold(prop.Name, propToExtract) {
				propValue = prop.Value
				break
			}
		}
		if propValue != "" {
			break
		}
	}

	if propValue == "" {
		return "", fmt.Errorf("Could not find value for '%s', xml: %+v", propToExtract, responseXML)
	}

	return propValue, nil
}

func (w *winHelper) extractSingleIntValue(propToExtract string, wmicArgs []string) (int, error) {
	valStr, err := w.extractSingleValue(propToExtract, wmicArgs)
	if err != nil {
		return 0, fmt.Errorf("Could not get %s, error: %s", propToExtract, err.Error())
	}
	valInt, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Could not parse %s '%s' as int, error: %s", propToExtract, valStr, err.Error())
	}
	return int(valInt), nil
}

func (w *winHelper) runExec(cmdLine ...string) ([]byte, error) {
	out, err := exec.Command(cmdLine[0], cmdLine[1:]...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Cannot run command (cmdline '%s'), error: %s", strings.Join(cmdLine, " "), err.Error())
	}
	return out, nil
}

func (w *winHelper) stripNonNumericCharsAndCastInt(s string) (int, error) {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			buf.WriteByte(s[i])
		}
	}

	str := buf.String()
	intVal, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Cannot parse int from string '%s', error: %s", str, err.Error())
	}
	return int(intVal), nil
}

func (w *winHelper) getCPUTotalDurationFromCsvString(s string) (time.Duration, error) {
	split := strings.Split(s, ":") //Expected format is HH:MM:SS
	if len(split) != 3 {
		return 0, fmt.Errorf("Unexpected cpu time string '%s' (must be in format HH:MM:SS)", s)
	}
	hourStr := split[0]
	minStr := split[1]
	secStr := split[2]

	hour, err := w.stripNonNumericCharsAndCastInt(hourStr)
	if err != nil {
		return 0, fmt.Errorf("Cannot cast hour value, error: %s", err.Error())
	}
	min, err := w.stripNonNumericCharsAndCastInt(minStr)
	if err != nil {
		return 0, fmt.Errorf("Cannot cast minute value, error: %s", err.Error())
	}
	sec, err := w.stripNonNumericCharsAndCastInt(secStr)
	if err != nil {
		return 0, fmt.Errorf("Cannot cast second value, error: %s", err.Error())
	}

	return time.Duration(sec)*time.Second + time.Duration(min)*time.Minute + time.Duration(hour)*time.Hour, nil
}

func (w *winHelper) CPUPercentage() (int, error) {
	return w.extractSingleIntValue("loadpercentage", []string{"cpu", "get", "loadpercentage"})
}

func (w *winHelper) FreePhysicalMemoryKB() (int, error) {
	return w.extractSingleIntValue("freephysicalmemory", []string{"os", "get", "freephysicalmemory"})
}

func (w *winHelper) FreeVirtualMemoryKB() (int, error) {
	return w.extractSingleIntValue("freevirtualmemory", []string{"os", "get", "freevirtualmemory"})
}

func (w *winHelper) ProcessUsedCPUAndMemoryKB(pid int) (memKB int, cpuDuration time.Duration, returnErr error) {
	// tasklist /FI "PID eq 46392" /FO csv
	csvResponse, err := w.runExec("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "csv", "/V")
	if err != nil {
		return 0, 0, err
	}

	type cell struct { // Our example struct, you can use "-" to ignore a field
		ImageName   string `csv:"Image Name"`
		PID         int    `csv:"PID"`
		SessionName string `csv:"Session Name"`
		SessionNum  string `csv:"Session#"`
		MemUsage    string `csv:"Mem Usage"`
		CPUTime     string `csv:"CPU Time"`
	}
	rows := []*cell{}
	in := strings.NewReader(string(csvResponse))
	err = gocsv.Unmarshal(in, &rows)
	if err != nil {
		return 0, 0, fmt.Errorf("Cannot unmarshal csv. Error: %s. CSV response was: %s", err.Error(), string(csvResponse))
	}

	tmpstrs := []string{}
	for _, r := range rows {
		tmpstrs = append(tmpstrs, fmt.Sprintf("%+v", r))
	}

	if len(rows) == 0 {
		return 0, 0, fmt.Errorf("Invalid csv data (no rows). CSV response was: %s", string(csvResponse))
	}
	row := rows[0]

	memUsageKB, err := w.stripNonNumericCharsAndCastInt(row.MemUsage)
	if err != nil {
		return 0, 0, fmt.Errorf("Cannot get process mem usage from csv data. Error: %s. CSV response was: %s", err.Error(), string(csvResponse))
	}

	cpuTimeDuration, err := w.getCPUTotalDurationFromCsvString(row.CPUTime)
	if err != nil {
		return 0, 0, fmt.Errorf("Unable to determine CPU Time (from string '%s'), error: %s", row.CPUTime, err.Error())
	}

	return memUsageKB, cpuTimeDuration, nil
}
