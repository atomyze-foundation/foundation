package version

import "os"

// CoreChaincodeIDName returns the core chaincode ID name
func CoreChaincodeIDName() string {
	ch := os.Getenv("CORE_CHAINCODE_ID_NAME")
	if ch == "" {
		return "'CORE_CHAINCODE_ID_NAME' is empty"
	}

	return ch
}
