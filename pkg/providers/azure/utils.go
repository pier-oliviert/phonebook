package azure

import (
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

// Helper function to convert string to *int32
func toInt32Ptr(s string) *int32 {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return to.Ptr(int32(i))
}
