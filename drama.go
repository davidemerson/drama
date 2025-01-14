package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ports = map[string]int{
	"IPP": 631,      // Internet Printing Protocol
	"AirPrint": 5353, // AirPrint/Multicast DNS
	"RAW": 9100,      // RAW printing protocol
	"LPD": 515,       // Line Printer Daemon
}

func scanPort(host string, port int, protocol string, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return
	}
	conn.Close()
	results <- fmt.Sprintf("%s,%s", host, protocol)
}

func getRandomPDF(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("could not read directory: %w", err)
	}

	pdfFiles := []string{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pdf" {
			pdfFiles = append(pdfFiles, filepath.Join(dir, file.Name()))
		}
	}

	if len(pdfFiles) == 0 {
		return "", fmt.Errorf("no PDF files found in directory")
	}

	return pdfFiles[rand.Intn(len(pdfFiles))], nil
}

func sendToPrinterWithLP(printerName, filePath string) error {
	cmd := exec.Command("lp", "-d", printerName, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send print job: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("Print job submitted successfully. Output: %s\n", string(output))
	return nil
}

func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	if ip == nil {
		return nil
	}
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
	return ip
}

func main() {
	logFile, err := os.Create("printer_scan_and_print.log")
	if err != nil {
		log.Fatalf("Could not create log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter IP range in CIDR notation (e.g., 192.168.1.0/24): ")
	cidr, _ := reader.ReadString('\n')
	cidr = strings.TrimSpace(cidr)

	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		logger.Fatalf("Invalid CIDR notation: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan string, 100)
	uniqueIPs := make(map[string]map[string]bool)
	var mu sync.Mutex

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); ip = nextIP(ip) {
		for protocol, port := range ports {
			wg.Add(1)
			go scanPort(ip.String(), port, protocol, &wg, results)
		}
	}

	go func() {
		for result := range results {
			fields := strings.Split(result, ",")
			ip, protocol := fields[0], fields[1]
			mu.Lock()
			if _, exists := uniqueIPs[ip]; !exists {
				uniqueIPs[ip] = make(map[string]bool)
			}
			uniqueIPs[ip][protocol] = true
			mu.Unlock()
		}
	}()

	wg.Wait()
	close(results)

	printers := []string{}
	for ip := range uniqueIPs {
		printers = append(printers, ip)
	}

	fmt.Println("Discovered printers:")
	for _, printer := range printers {
		fmt.Println(printer)
	}

	for _, printer := range printers {
		fmt.Printf("Confirm sending print job to %s (y/n): ", printer)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if strings.ToLower(response) != "y" {
			continue
		}

		filePath, err := getRandomPDF(".")
		if err != nil {
			logger.Printf("Failed to get a random PDF: %v", err)
			continue
		}

		logger.Printf("Attempting to send %s to printer at %s", filePath, printer)
		if err := sendToPrinterWithLP(printer, filePath); err != nil {
			logger.Printf("Failed to send PDF to printer at %s: %v", printer, err)
		} else {
			logger.Printf("Successfully sent %s to printer at %s", filePath, printer)
		}
	}

	logger.Println("Execution completed.")
}
