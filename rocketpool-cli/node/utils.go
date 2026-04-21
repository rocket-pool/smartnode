package node

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

// IPInfo API
const IPInfoURL = "https://ipinfo.io/json/"

// IPInfo response
type ipInfoResponse struct {
	Timezone string `json:"timezone"`
}

// Prompt user for a time zone string
func promptTimezone() string {

	// Time zone value
	var timezone string

	// Prompt for auto-detect
	if prompt.Confirm("Would you like to detect your timezone automatically?") {
		// Detect using the IPInfo API
		resp, err := http.Get(IPInfoURL)
		if err == nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				message := new(ipInfoResponse)
				err := json.Unmarshal(body, message)
				if err == nil {
					timezone = message.Timezone
				} else {
					fmt.Printf("WARNING: couldn't query %s for your timezone based on your IP address (%s).\nChecking your system's timezone...\n", IPInfoURL, err.Error())
				}
			} else {
				fmt.Printf("WARNING: couldn't query %s for your timezone based on your IP address (%s).\nChecking your system's timezone...\n", IPInfoURL, err.Error())
			}
		} else {
			fmt.Printf("WARNING: couldn't query %s for your timezone based on your IP address (%s).\nChecking your system's timezone...\n", IPInfoURL, err.Error())
		}

		// Fall back to system time zone
		if timezone == "" {
			_, err := os.Stat("/etc/timezone")
			if os.IsNotExist(err) {
				// Try /etc/localtime, which Redhat-based systems use instead
				_, err = os.Stat("/etc/localtime")
				if err != nil {
					fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
				} else {
					path, err := filepath.EvalSymlinks("/etc/localtime")
					if err != nil {
						fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
					} else {
						path = strings.TrimPrefix(path, "/usr/share/zoneinfo/")
						path = strings.TrimPrefix(path, "posix/")
						path = strings.TrimSpace(path)
						// Verify it
						_, err = time.LoadLocation(path)
						if err != nil {
							fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
						} else {
							timezone = path
						}
					}
				}
			} else if err != nil {
				fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
			} else {
				// Debian systems
				bytes, err := os.ReadFile("/etc/timezone")
				if err != nil {
					fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
				} else {
					timezone = strings.TrimSpace(string(bytes))
					// Verify it
					_, err = time.LoadLocation(timezone)
					if err != nil {
						fmt.Printf("WARNING: couldn't get system timezone info (%s), you'll have to set it manually.\n", err.Error())
						timezone = ""
					}
				}
			}
		}

	}

	// Confirm detected time zone
	if timezone != "" {
		if !prompt.Confirm("The detected timezone is '%s', would you like to register using this timezone?", timezone) {
			timezone = ""
		} else {
			return timezone
		}
	}

	// Get the list of valid countries
	platformZoneSources := []string{
		"/usr/share/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		"/usr/lib/locale/TZ/",
	}
	invalidCountries := []string{
		"SystemV",
	}

	countryNames := []string{}
	for _, source := range platformZoneSources {
		files, err := os.ReadDir(source)
		if err != nil {
			continue
		}

		for _, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			filename := fileInfo.Name()
			isSymlink := fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink // Don't allow symlinks, which are just TZ aliases
			isDir := fileInfo.IsDir()                                     // Must be a directory
			isUpper := unicode.IsUpper(rune(filename[0]))                 // Must start with an upper case letter
			if !isSymlink && isDir && isUpper {
				isValid := true
				if slices.Contains(invalidCountries, filename) {
					isValid = false
				}
				if isValid {
					countryNames = append(countryNames, filename)
				}
			}
		}
	}

	fmt.Println("You will now be prompted to enter a timezone.")
	fmt.Println("For a complete list of valid entries, please use one of the \"TZ database name\" entries listed here:")
	fmt.Println("https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
	fmt.Println()

	// Handle situations where we couldn't parse any timezone info from the OS
	if len(countryNames) == 0 {
		for timezone == "" {
			timezone = prompt.Prompt("Please enter a timezone to register with in the format 'Country/City' (use Etc/UTC if you prefer not to answer):", "^([a-zA-Z_]{2,}\\/)+[a-zA-Z_]{2,}$", "Please enter a timezone in the format 'Country/City' (use Etc/UTC if you prefer not to answer)")
			if !prompt.Confirm("You have chosen to register with the timezone '%s', is this correct?", timezone) {
				timezone = ""
			}
		}

		// Return
		return timezone
	}

	// Print countries
	sort.Strings(countryNames)
	fmt.Println("List of valid countries / continents:")
	for _, countryName := range countryNames {
		fmt.Println(countryName)
	}
	fmt.Println()

	// Prompt for country
	country := ""
	for {
		country = prompt.Prompt("Please enter a country / continent from the list above:", "^.+$", "Please enter a country / continent from the list above:")

		exists := slices.Contains(countryNames, country)

		if !exists {
			fmt.Printf("%s is not a valid country or continent. Please see the list above for valid countries and continents.\n\n", country)
		} else {
			break
		}
	}

	// Get the list of regions for the selected country
	regionNames := []string{}
	for _, source := range platformZoneSources {
		files, err := os.ReadDir(filepath.Join(source, country))
		if err != nil {
			continue
		}

		for _, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			if fileInfo.IsDir() {
				subfiles, err := os.ReadDir(filepath.Join(source, country, fileInfo.Name()))
				if err != nil {
					continue
				}
				for _, subfile := range subfiles {
					subfileInfo, err := subfile.Info()
					if err != nil {
						continue
					}
					regionNames = append(regionNames, fmt.Sprintf("%s/%s", fileInfo.Name(), subfileInfo.Name()))
				}
			} else {
				regionNames = append(regionNames, fileInfo.Name())
			}
		}
	}

	// Print regions
	sort.Strings(regionNames)
	fmt.Println("List of valid regions:")
	for _, regionName := range regionNames {
		fmt.Println(regionName)
	}
	fmt.Println()

	// Prompt for region
	region := ""
	for {
		region = prompt.Prompt("Please enter a region from the list above:", "^.+$", "Please enter a region from the list above:")

		exists := slices.Contains(regionNames, region)

		if !exists {
			fmt.Printf("%s is not a valid country or continent. Please see the list above for valid countries and continents.\n\n", region)
		} else {
			break
		}
	}

	// Return
	timezone = fmt.Sprintf("%s/%s", country, region)
	fmt.Printf("Using timezone %s.\n", timezone)
	return timezone
}
