package updatecheck

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type autoUpdateChecker interface {
	Run() error
}

type autoUpdateCheckerImpl struct {
	imageId         string
	reusultFilePath string
}

func NewAutoUpdateChecker(imageId string, resultFilePath string) autoUpdateChecker {
	return &autoUpdateCheckerImpl{
		imageId:         imageId,
		reusultFilePath: resultFilePath,
	}
}

func (auci *autoUpdateCheckerImpl) Run() error {
	log.Printf("Running update check for image id '%s'\n", auci.imageId)
	if pathExists("/etc/alpine-release") {

		log.Println("Detected /etc/alpine-release, assuming alpine linux, checking if apk exist")
		apkPath, err := exec.LookPath("apk")
		if err != nil {
			return fmt.Errorf("could not find apk executable, wrapped error is: '%w'", err)
		}
		log.Printf("Found apk executable in '%s'\n", apkPath)

		return performAlpineUpdateCheck(auci.reusultFilePath)
	} else if pathExists("/etc/redhat-release") {
		log.Println("Detected /etc/redhat-release, assuming rpm based distribution, checking if yum or microdnf and rpm exists")

		yumPath, err := exec.LookPath("yum")
		if err != nil {
			log.Printf("Did not find yum (error is '%v'), looking for microdnf\n", err)
			microdnfPath, err := exec.LookPath("microdnf")
			if err != nil {
				return fmt.Errorf("did not find microdnf (error is '%v'), cannot continue", err)
			} else {
				log.Printf("Found microdnf executable in '%s'\n", microdnfPath)
				return performRedHatUpdateCheck("microdnf", auci.reusultFilePath)
			}
		} else {
			log.Printf("Found yum executable in '%s'\n", yumPath)
			return performRedHatUpdateCheck("yum", auci.reusultFilePath)
		}

	} else if pathExists("/etc/debian_version") {
		log.Println("Detected /etc/debian_version, assuming debian based distribution, checking apt-get and dpkg exists")
		aptGetPath, err := exec.LookPath("apt-get")
		if err != nil {
			return fmt.Errorf("could not find apt-get executable, wrapped error is: '%w'", err)
		}
		log.Printf("Found apt-get executable in '%s'\n", aptGetPath)
		return performDebianUpdateCheck(auci.reusultFilePath)
	}
	return nil
}

func performUpdateCheckCommand(packageListCmd []string, updateCommands [][]string, resultFile string) {
	log.Printf("Performing package list command before upgrade '%s'\n", strings.Join(packageListCmd, " "))
	pkgListCmd := exec.Command(packageListCmd[0], packageListCmd[1:]...) // #nosec G204
	pkgListCmd.Stderr = os.Stderr
	output, err := pkgListCmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	beforeUpgradeList := strings.Split(string(output), "\n")
	sort.Strings(beforeUpgradeList)

	for _, updateCmd := range updateCommands {
		log.Printf("Performing update check command '%s'\n", strings.Join(updateCmd, " "))
		updateCommand := exec.Command(updateCmd[0], updateCmd[1:]...) // #nosec G204
		updateCommand.Stderr = os.Stderr
		updateCommand.Stdout = os.Stdout
		err := updateCommand.Run()
		if err != nil {
			panic(err)
		}
	}
	log.Printf("Performing package list command after upgrade '%s'\n", strings.Join(packageListCmd, " "))
	pkgListCmd = exec.Command(packageListCmd[0], packageListCmd[1:]...) // #nosec G204
	pkgListCmd.Stderr = os.Stderr
	output, err = pkgListCmd.Output()
	if err != nil {
		panic(err)
	}
	afterUpgradeList := strings.Split(string(output), "\n")
	sort.Strings(afterUpgradeList)

	log.Println("Comparing package list from before and after upgrade")
	if stringSliceEquals(beforeUpgradeList, afterUpgradeList) {
		log.Printf("both package lists equal, writing NO_UPGRADE_NEEDED to '%s'\n", resultFile)
		err := os.WriteFile(resultFile, []byte("NO_UPGRADE_NEEDED\n"), 0600)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("package lists do not equal, that indicates an update, writing UPGRADE_NEEDED to '%s'\n", resultFile)
		err := os.WriteFile(resultFile, []byte("UPGRADE_NEEDED\n"), 0600)
		if err != nil {
			panic(err)
		}
	}
}

func stringSliceEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for idx, elem := range a {
		if b[idx] != elem {
			return false
		}
	}
	return true
}

func performRedHatUpdateCheck(microdnfOrYum string, resultFilePath string) error {
	performUpdateCheckCommand([]string{"rpm", "-qa"}, [][]string{{microdnfOrYum, "-y", "upgrade"}}, resultFilePath)
	return nil
}

func performDebianUpdateCheck(resultFilePath string) error {
	performUpdateCheckCommand([]string{"dpkg", "-l"}, [][]string{{"apt-get", "update"}, {"apt-get", "-y", "dist-upgrade"}}, resultFilePath)
	return nil
}

func performAlpineUpdateCheck(resultFilePath string) error {
	performUpdateCheckCommand([]string{"apk", "info", "-v"}, [][]string{{"apk", "upgrade"}}, resultFilePath)
	return nil
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}
