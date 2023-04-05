package trigger

import "os"

const (
	FileName  = "/Users/jamesmilligan/code/flagd-1/config/samples/example_flags.json"
	StartSpec = `
	{
		"flags": {
		  "myBoolFlag": {
			"state": "ENABLED",
			"variants": {
			  "on": true,
			  "off": false
			},
			"defaultVariant": "off"
		  }
		}
	}`
	UpdatedSpec = `
	{
		"flags": {
		  "myBoolFlag": {
			"state": "ENABLED",
			"variants": {
			  "on": true,
			  "off": false
			},
			"defaultVariant": "on"
		  }
		} 
	}`
)

func SetupFile() error {
	return os.WriteFile(FileName, []byte(StartSpec), 0644)
}

func UpdateFile() error {
	return os.WriteFile(FileName, []byte(UpdatedSpec), 0644)
}

func Cleanup() error {
	return os.Remove(FileName)
}
