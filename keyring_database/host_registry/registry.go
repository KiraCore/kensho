package host_registry

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type HostRegistry struct {
	filePath string
	mu       sync.Mutex
	ips      map[string]struct{}
}

func NewIPRegistry(filePath string) (*HostRegistry, error) {
	registry := &HostRegistry{
		filePath: filePath,
		ips:      make(map[string]struct{}),
	}

	// Load existing IPs from file
	err := registry.loadFromFile()
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (r *HostRegistry) loadFromFile() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			r.ips[ip] = struct{}{}
		}
	}

	return scanner.Err()
}

func (r *HostRegistry) saveToFile() error {
	file, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for ip := range r.ips {
		_, err := file.WriteString(ip + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *HostRegistry) AddIP(ip string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.ips[ip]; exists {
		return nil
	}

	r.ips[ip] = struct{}{}
	return r.saveToFile()
}

func (r *HostRegistry) DeleteIP(ip string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.ips[ip]; exists {
		delete(r.ips, ip)
		return r.saveToFile()
	}
	return fmt.Errorf("IP %s not found", ip)
}

func (r *HostRegistry) ListIPs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []string
	for ip := range r.ips {
		result = append(result, ip)
	}

	return result
}
