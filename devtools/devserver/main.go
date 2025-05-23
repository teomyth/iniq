package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Parse command line arguments
	port := flag.Int("port", 8080, "Port to listen on")
	binDir := flag.String("bin", "bin", "Directory containing binaries")
	flag.Parse()

	// Resolve absolute paths
	absBinDir, err := filepath.Abs(*binDir)
	if err != nil {
		log.Fatalf("Failed to resolve bin directory: %v", err)
	}

	// Check if directories exist
	if _, err := os.Stat(absBinDir); os.IsNotExist(err) {
		log.Printf("Bin directory does not exist: %s. Creating it...", absBinDir)
		if err := os.MkdirAll(absBinDir, 0755); err != nil {
			log.Fatalf("Failed to create bin directory: %v", err)
		}
	}

	// Ensure scripts are built
	if _, err := os.Stat(filepath.Join(absBinDir, "install.sh")); os.IsNotExist(err) {
		log.Printf("Warning: install.sh not found in bin directory. Run 'task build:scripts' first.")
	}

	// Set up HTTP routes
	http.HandleFunc("/install.sh", func(w http.ResponseWriter, r *http.Request) {
		serveScriptWithDynamicValues(w, r, filepath.Join(*binDir, "install.sh"))
	})

	// Handle binary downloads and static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a binary download request
		if strings.HasPrefix(r.URL.Path, "/iniq-") {
			// Extract the requested binary name from the URL path
			requestedName := filepath.Base(r.URL.Path)

			// Check if this is a tar.gz request
			isTarGzRequest := strings.HasSuffix(requestedName, ".tar.gz")

			// Only log the binary request, not the full directory listing
			log.Printf("Binary download request: %s", r.URL.Path)

			var baseBinaryName, baseBinaryPath string
			if isTarGzRequest {
				// For tar.gz requests, get the binary name without .tar.gz extension
				baseBinaryName = strings.TrimSuffix(requestedName, ".tar.gz")
				baseBinaryPath = filepath.Join(absBinDir, baseBinaryName)
			} else {
				// Direct binary request
				baseBinaryName = requestedName
				baseBinaryPath = filepath.Join(absBinDir, baseBinaryName)
			}

			// Check if the base binary exists
			if _, err := os.Stat(baseBinaryPath); os.IsNotExist(err) {
				log.Printf("Base binary not found: %s", baseBinaryPath)
				http.Error(w, "Base binary not found", http.StatusNotFound)
				return
			}

			if isTarGzRequest {
				// Generate tar.gz on-the-fly
				log.Printf("Generating tar.gz for: %s", baseBinaryPath)

				w.Header().Set("Content-Type", "application/gzip")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", requestedName))

				// Create gzip writer
				gzipWriter := gzip.NewWriter(w)
				defer gzipWriter.Close()

				// Create tar writer
				tarWriter := tar.NewWriter(gzipWriter)
				defer tarWriter.Close()

				// Get file info
				fileInfo, err := os.Stat(baseBinaryPath)
				if err != nil {
					log.Printf("Error getting file info: %v", err)
					http.Error(w, "Error getting file info", http.StatusInternalServerError)
					return
				}

				// Create tar header
				header := &tar.Header{
					Name:    "iniq", // Always name the binary "iniq" inside the archive
					Mode:    0755,
					Size:    fileInfo.Size(),
					ModTime: fileInfo.ModTime(),
				}

				// Write header
				if err := tarWriter.WriteHeader(header); err != nil {
					log.Printf("Error writing tar header: %v", err)
					http.Error(w, "Error writing tar header", http.StatusInternalServerError)
					return
				}

				// Copy file content
				file, err := os.Open(baseBinaryPath)
				if err != nil {
					log.Printf("Error opening file: %v", err)
					http.Error(w, "Error opening file", http.StatusInternalServerError)
					return
				}
				defer file.Close()

				if _, err := io.Copy(tarWriter, file); err != nil {
					log.Printf("Error copying file content: %v", err)
					http.Error(w, "Error copying file content", http.StatusInternalServerError)
					return
				}

				log.Printf("Successfully served tar.gz: %s", requestedName)
				return
			} else {
				// Serve direct binary file
				log.Printf("Serving binary file: %s", baseBinaryPath)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", baseBinaryName))
				http.ServeFile(w, r, baseBinaryPath)
				return
			}
		}

		// If not a binary download and not root path, check for static files
		if r.URL.Path != "/" {
			// Try to serve static files
			staticFilePath := filepath.Join("devtools", "devserver", "static", filepath.Clean(r.URL.Path))
			if _, err := os.Stat(staticFilePath); err == nil {
				http.ServeFile(w, r, staticFilePath)
				return
			}

			// If not a static file, return 404
			http.NotFound(w, r)
			return
		}

		// Handle root path - serve the index.html file with dynamic replacements
		indexPath := filepath.Join("devtools", "devserver", "static", "index.html")

		// Check if index.html exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			log.Printf("Index file not found: %s", indexPath)
			http.Error(w, "Index file not found", http.StatusInternalServerError)
			return
		}

		// Read the index.html file
		content, err := os.ReadFile(indexPath)
		if err != nil {
			log.Printf("Error reading index file: %v", err)
			http.Error(w, "Error reading index file", http.StatusInternalServerError)
			return
		}

		// Replace placeholders with dynamic values
		serverURL := getServerURL(r)
		htmlContent := strings.ReplaceAll(string(content), "{{SERVER_URL}}", serverURL)

		// Serve the modified HTML
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(htmlContent))
	})

	// Status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","server":"INIQ Development Server"}`)
	})

	// Health check endpoint for heartbeat
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Note: Root endpoint is now handled in the combined handler above

	// Start server
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Starting development server on %s", addr)
	log.Printf("Bin directory: %s", absBinDir)

	// Print all available IP addresses
	printAvailableAddresses(*port)

	log.Fatal(http.ListenAndServe(addr, nil))
}

// Get server URL from request
func getServerURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	// Special case for cloudflared tunnels
	if strings.Contains(r.Host, "trycloudflare.com") {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// Serve script files with dynamic value replacement
func serveScriptWithDynamicValues(w http.ResponseWriter, r *http.Request, scriptPath string) {
	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		http.Error(w, "Script not found", http.StatusNotFound)
		return
	}

	// Read file content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		http.Error(w, "Failed to read script", http.StatusInternalServerError)
		return
	}

	// Get server URL
	serverURL := getServerURL(r)

	// Replace dynamic values in the script
	modifiedContent := string(content)

	// We don't need to replace REQUEST_HOST and REQUEST_PROTOCOL anymore
	// as we now use the full server URL directly in DEFAULT_DOWNLOAD_BASE_URL

	// Replace download URLs with the full server URL (including protocol)
	urlRegex := regexp.MustCompile(`DEFAULT_DOWNLOAD_BASE_URL="[^"]*"`)

	// Always use the full server URL with protocol
	modifiedContent = urlRegex.ReplaceAllString(modifiedContent, fmt.Sprintf(`DEFAULT_DOWNLOAD_BASE_URL="%s"`, serverURL))

	// Set content type and return script
	w.Header().Set("Content-Type", "text/plain")
	_, _ = io.WriteString(w, modifiedContent)
}

// Print all available IP addresses
func printAvailableAddresses(port int) {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error getting network interfaces: %v", err)
		return
	}

	// Print header
	fmt.Println("\n🌐 Access URLs:")

	// Always include 127.0.0.1
	fmt.Printf("  http://127.0.0.1:\033[0;32m%d\033[0m  (Local loopback)\n", port)

	// Track if we've found any important addresses
	foundImportant := false

	// Process each interface
	for _, iface := range interfaces {
		// Skip loopback, down, and non-multicast interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Process each address
		for _, addr := range addrs {
			// Extract IP address
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			// Skip IPv6 addresses for simplicity
			if ip.To4() == nil {
				continue
			}

			// Skip loopback addresses (already handled)
			if ip.IsLoopback() {
				continue
			}

			// Identify network type
			networkType := identifyNetworkType(ip.String(), iface.Name)

			// Format the output
			if networkType == "WSL Host Network (Important)" || networkType == "Physical Network (Important)" {
				// Highlight important addresses
				fmt.Printf("  \033[1;32mhttp://%s:\033[0;32m%d\033[1;32m  (%s)\033[0m\n", ip.String(), port, networkType)
				foundImportant = true
			} else {
				fmt.Printf("  http://%s:\033[0;32m%d\033[0m  (%s)\n", ip.String(), port, networkType)
			}
		}
	}

	// Add a note if no important addresses were found
	if !foundImportant {
		fmt.Println("\n⚠️  No important network addresses found. You may have connectivity issues.")
	}

	fmt.Println("\n💡 Use the highlighted address for external access.")
	fmt.Println("   For WSL, use the 'WSL Host Network' address from your Windows host.")
	fmt.Println()
}

// Identify the type of network based on IP address and interface name
func identifyNetworkType(ipAddr string, ifaceName string) string {
	// Check for common WSL interface names
	if strings.Contains(ifaceName, "eth") {
		// Check for WSL specific IP ranges
		if strings.HasPrefix(ipAddr, "172.") {
			return "WSL Internal Network"
		}
	}

	// Check for Docker interface
	if strings.Contains(ifaceName, "docker") || strings.Contains(ifaceName, "br-") {
		return "Docker Network"
	}

	// Check for VPN interfaces
	if strings.Contains(ifaceName, "tun") || strings.Contains(ifaceName, "tap") {
		return "VPN Network"
	}

	// Check for WSL host network (important)
	if strings.HasPrefix(ipAddr, "10.1.1.") {
		return "WSL Host Network (Important)"
	}

	// Check for WSL NAT network
	if strings.HasPrefix(ipAddr, "10.255.255.") {
		return "WSL NAT Network"
	}

	// Check for WSL/Docker shared network
	if strings.HasPrefix(ipAddr, "100.64.") {
		return "WSL/Docker Shared Network"
	}

	// Check for common private network ranges
	if strings.HasPrefix(ipAddr, "192.168.") ||
		strings.HasPrefix(ipAddr, "10.") ||
		(strings.HasPrefix(ipAddr, "172.") &&
			strings.Split(ipAddr, ".")[1] >= "16" &&
			strings.Split(ipAddr, ".")[1] <= "31") {
		return "Physical Network (Important)"
	}

	// Default case
	return "Unknown Network"
}

// Note: This function is kept for reference but not currently used
// as we're using serveScriptWithDynamicValues instead
/*
func serveScript(w http.ResponseWriter, r *http.Request, scriptPath string) {
	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		http.Error(w, "Script not found", http.StatusNotFound)
		return
	}

	// Read file content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		http.Error(w, "Failed to read script", http.StatusInternalServerError)
		return
	}

	// Set content type and return script
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}
*/
