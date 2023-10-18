package version

import "os"

func CoreChaincodeIDName() string {
	ch := os.Getenv("CORE_CHAINCODE_ID_NAME")
	if ch == "" {
		return "'CORE_CHAINCODE_ID_NAME' is empty"
	}

	return ch
}
