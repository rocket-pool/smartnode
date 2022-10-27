package sys

import (
	"runtime"
	"sort"

	"github.com/klauspost/cpuid/v2"
)

// Returns the CPU features that are required to run "modern" images but are not present on the node's CPU
func GetMissingModernCpuFeatures() []string {
	var features map[cpuid.FeatureID]string
	switch runtime.GOARCH {
	case "amd64":
		features = map[cpuid.FeatureID]string{
			cpuid.ADX: "adx",
			//cpuid.AESNI: "aes",
			cpuid.AVX:   "avx",
			cpuid.AVX2:  "avx2",
			cpuid.BMI1:  "bmi1",
			cpuid.BMI2:  "bmi2",
			cpuid.CLMUL: "clmul",
			cpuid.MMX:   "mmx",
			cpuid.SSE:   "sse",
			cpuid.SSE2:  "sse2",
			cpuid.SSSE3: "ssse3",
			cpuid.SSE4:  "sse4.1",
			cpuid.SSE42: "sse4.2",
		}
	default:
		features = map[cpuid.FeatureID]string{}
	}

	unsupportedFeatures := []string{}
	for feature, name := range features {
		if !cpuid.CPU.Supports(feature) {
			unsupportedFeatures = append(unsupportedFeatures, name)
		}
	}

	sort.Strings(unsupportedFeatures)
	return unsupportedFeatures
}
