package update

import (
	"fmt"
	"strconv"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

// writeCounter tracks download progress
type WriteCounter struct {
	Total int64
	count int64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.count += int64(n)
	wc.printProgress()
	return n, nil
}

func (wc *WriteCounter) printProgress() {
	// If we don't know the total size, just show downloaded bytes
	if wc.Total <= 0 {
		logger.Backup.Info(fmt.Sprintf("\r%s downloaded", bytesToHuman(wc.count)))
		return
	}

	// Calculate percentage with bounds checking
	percent := float64(wc.count) / float64(wc.Total) * 100
	if percent > 100 {
		percent = 100
	}

	// Create simple progress bar
	width := 20
	complete := int(percent / 100 * float64(width))

	progressBar := "["
	for i := 0; i < width; i++ {
		if i < complete {
			progressBar += "="
		} else if i == complete && complete < width {
			progressBar += ">"
		} else {
			progressBar += " "
		}
	}
	progressBar += "]"

	// Print progress and erase to end of line
	logger.Backup.Info(fmt.Sprintf("\r%s %.1f%% (%s/%s)",
		progressBar,
		percent,
		bytesToHuman(wc.count),
		bytesToHuman(wc.Total)))
}

// bytesToHuman converts bytes to human readable format
func bytesToHuman(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
