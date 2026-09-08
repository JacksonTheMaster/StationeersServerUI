// customdetections.go

package detectionmgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

var ErrDetectionNotFound = errors.New("detection not found")

var customDetectionsManager *CustomDetectionsManager

/*
Custom Detection Pattern Management System
- Manages lifecycle of user-defined detection patterns (regex/keywords)
- Provides CRUD operations for patterns persisted in JSON format
- Handles pattern validation, storage synchronization, and detector updates
- Thread-safe implementation with RW mutex for concurrent access
- Integrated with main Detector for real-time pattern application
*/

// CustomDetectionsManager handles loading, saving and managing custom detections
type CustomDetectionsManager struct {
	Detections []CustomDetection
	detector   *Detector
	mutex      sync.RWMutex
}

// CustomDetection defines a user-defined detection pattern
type CustomDetection struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "regex" or "keyword"
	Pattern   string `json:"pattern"`
	EventType string `json:"eventType"`
	Message   string `json:"message"`
}

// NewCustomDetectionsManager creates a new manager and loads existing detections
func NewCustomDetectionsManager(detector *Detector) *CustomDetectionsManager {
	manager := &CustomDetectionsManager{
		Detections: []CustomDetection{},
		detector:   detector,
	}
	manager.LoadDetections()
	return manager
}

func InitCustomDetectionsManager(detector *Detector) {
	customDetectionsManager = NewCustomDetectionsManager(detector)
}

func GetCustomDetections() []CustomDetection {
	if customDetectionsManager == nil {
		return []CustomDetection{}
	}
	return customDetectionsManager.GetDetections()
}

func AddCustomDetection(detection CustomDetection) error {
	if customDetectionsManager == nil {
		return errors.New("custom detections manager is not initialized")
	}
	return customDetectionsManager.AddDetection(detection)
}

func RemoveCustomDetection(id string) error {
	if customDetectionsManager == nil {
		return errors.New("custom detections manager is not initialized")
	}
	return customDetectionsManager.DeleteDetection(id)
}

// LoadDetections loads custom detections from file
func (m *CustomDetectionsManager) LoadDetections() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(config.GetCustomDetectionsFilePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists, create if not
	if _, err := os.Stat(config.GetCustomDetectionsFilePath()); os.IsNotExist(err) {
		// Create empty file
		file, err := os.Create(config.GetCustomDetectionsFilePath())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		file.Write([]byte("[]"))
		file.Close()
		return nil
	}

	// Read file
	file, err := os.Open(config.GetCustomDetectionsFilePath())
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode JSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&m.Detections); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Update detector
	m.updateDetector()
	return nil
}

// AddDetection adds a new custom detection
func (m *CustomDetectionsManager) AddDetection(detection CustomDetection) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate pattern
	if detection.Type == "regex" {
		_, err := regexp.Compile(detection.Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	// Add detection
	m.Detections = append(m.Detections, detection)

	// Update detector
	m.updateDetector()

	// Save to file directly
	file, err := os.Create(config.GetCustomDetectionsFilePath())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m.Detections); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// DeleteDetection removes a custom detection by ID
func (m *CustomDetectionsManager) DeleteDetection(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Find detection
	for i, detection := range m.Detections {
		if detection.ID == id {
			// Remove detection
			m.Detections = append(m.Detections[:i], m.Detections[i+1:]...)

			// Update detector
			m.updateDetector()

			// Save to file directly
			file, err := os.Create(config.GetCustomDetectionsFilePath())
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(m.Detections); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}

			return nil
		}
	}

	return ErrDetectionNotFound
}

// GetDetections returns all custom detections
func (m *CustomDetectionsManager) GetDetections() []CustomDetection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return copy of detections
	detections := make([]CustomDetection, len(m.Detections))
	copy(detections, m.Detections)
	return detections
}

// updateDetector updates the detector with current custom patterns
func (m *CustomDetectionsManager) updateDetector() {
	customPatterns := []CustomPattern{}

	for _, cd := range m.Detections {
		if cd.Type == "regex" {
			re, err := regexp.Compile(cd.Pattern)
			if err == nil {
				customPatterns = append(customPatterns, CustomPattern{
					Pattern:     re,
					EventType:   EventType(cd.EventType),
					MessageTmpl: cd.Message,
					IsRegex:     true,
				})
			}
		} else {
			customPatterns = append(customPatterns, CustomPattern{
				Keyword:     cd.Pattern,
				EventType:   EventType(cd.EventType),
				MessageTmpl: cd.Message,
				IsRegex:     false,
			})
		}
	}

	m.detector.SetCustomPatterns(customPatterns)
}
