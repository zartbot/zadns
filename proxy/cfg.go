package proxy

import (
	"bufio"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func ReadCfg(filename string) map[string][]string {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := make(map[string][]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		record := strings.Split(line, "|")
		if len(record) != 2 {
			logrus.Fatal("invalid config found:", filename, "|", record)
		}

		key := strings.TrimSpace(record[0])
		values := strings.Split(strings.TrimSpace(record[1]), ",")
		if len(values) > 0 {
			for k, v := range values {
				values[k] = strings.TrimSpace(v)
				if net.ParseIP(values[k]) == nil {
					logrus.Fatal("invalid server address: ", v, "@", key)
				}
			}
			result[key] = values
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.Fatal(err)
	}
	return result
}

func ReadServerListCfg(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := make([]string, 0)
	for scanner.Scan() {
		server := strings.TrimSpace(scanner.Text())
		if server == "" {
			continue
		}
		if net.ParseIP(server) == nil {
			logrus.Fatal("invalid server address: ", server)
		}
		server = server + ":53"
		result = append(result, server)
	}
	if err := scanner.Err(); err != nil {
		logrus.Fatal(err)
	}
	return result
}
