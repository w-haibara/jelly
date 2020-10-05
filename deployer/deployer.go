package deployer

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Deploy projects
func Deploy(arg string) (string, error) {
	log.Info("Deploy starting")
	pwd := ""
	if out, err := exec.Command("pwd").Output(); err == nil {
		pwd = strings.TrimRight(string(out), "\n")
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Exec Command failed")
		return "", err
	}
	if arg == "latest" {
		cmd := exec.Command(pwd+"/script/deploy.sh", "latest")
		log.WithFields(log.Fields{
			"cmd:": cmd,
		}).Info("Execute")
		out, err := cmd.Output()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Command finished with error")
			return "", err
		}
		result := string(out)
		log.Info("Command finished")
		log.Println(result)
		return result, nil
	} else if len(arg) < 20 {
		cmd := exec.Command(pwd+"/script/deploy.sh", arg)
		log.WithFields(log.Fields{
			"cmd:": cmd,
		}).Info("Execute")
		out, err := cmd.Output()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Command finished with error")
			return "", err
		}
		result := string(out)
		log.Info("Command finished")
		log.Println(result)
		return result, nil
	}

	return "", nil
}
