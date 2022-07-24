package updatecheck

import "fmt"

func getImageUpdateCheckResultFileName(imageId string, locationIndex int) string {
	return fmt.Sprintf("gipgee-update-check-result-%s-release-location-%d", imageId, locationIndex)
}
