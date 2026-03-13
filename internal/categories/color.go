package categories

import "hash/fnv"

// ColorForExternalID returns a stable pseudo-random HSL color for a category.
func ColorForExternalID(externalID string) string {
	if externalID == "" {
		externalID = "default"
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(externalID))
	hash := hasher.Sum32()

	hue := hash % 360
	saturation := 50 + (hash % 10)
	lightness := 56 + ((hash >> 3) % 8)
	alpha := uint32(82)

	return hslString(hue, saturation, lightness, alpha)
}

func hslString(hue, saturation, lightness, alpha uint32) string {
	return "hsl(" +
		itoa(hue) +
		" " +
		itoa(saturation) +
		"% " +
		itoa(lightness) +
		"% / 0." +
		itoa(alpha) +
		")"
}

func itoa(v uint32) string {
	if v == 0 {
		return "0"
	}

	buf := [10]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}

	return string(buf[i:])
}
