package utils

import (
	"net"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetIPsFromRequest retrieves all IPs from various headers in the Gin context
func GetIPsFromRequest(c *gin.Context) string {
	IPsMap := make(map[string]struct{})

	// Get IP from RemoteAddr
	addIP(strings.Split(c.Request.RemoteAddr, ":")[0], IPsMap)

	// Get ClientIP from Gin's built-in method
	addIP(c.ClientIP(), IPsMap)

	// Get IPs from X-Forwarded-For and add them to the map
	for _, IP := range strings.Split(c.GetHeader("X-Forwarded-For"), ",") {
		addIP(IP, IPsMap)
	}

	// Get IP from X-Real-IP and add to map
	addIP(c.GetHeader("X-Real-IP"), IPsMap)

	// Get IP from Forwarded header
	forwardedHeader := c.GetHeader("Forwarded")
	for _, match := range regexp.MustCompile(`for=("[^"]*"|[^;,\s]+)`).FindAllStringSubmatch(forwardedHeader, -1) {
		if len(match) > 1 {
			// Remove potential double-quotes around the IP
			addIP(strings.Trim(match[1], `"`), IPsMap)
		}
	}

	// Construct a slice of IPs
	var ipsSlice []string
	for IP := range IPsMap {
		ipsSlice = append(ipsSlice, IP)
	}

	// Join all IPs as a single string
	return strings.Join(ipsSlice, ", ")
}

// addIP adds an IP to the map if it's valid and not already present
func addIP(IP string, IPsMap map[string]struct{}) {
	IP = strings.TrimSpace(IP)

	if IP != "" {
		_, IPNet, err := net.ParseCIDR(IP)
		if err == nil {
			IP = strings.Split(IPNet.String(), "/")[0]
		} else {
			// Could be just a plain IP without CIDR notation
			parsedIP := net.ParseIP(IP)
			if parsedIP != nil {
				IP = parsedIP.String()
			}
		}
		if _, ok := IPsMap[IP]; !ok {
			IPsMap[IP] = struct{}{}
		}
	}
}
