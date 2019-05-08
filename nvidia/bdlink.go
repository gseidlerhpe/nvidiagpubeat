//******************************************************************
//Copyright 2019 Hewlett Packard Corporation.
//Architect/Developer: Gernot Seidler

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at

// https://www.apache.org/licenses/LICENSE-2.0

//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//******************************************************************

package nvidia

import (
	"io"
	"os/exec"
	"strconv"
	"strings"
	"regexp"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

//BDLink_I provides interface to container name to GPU device links.
type BDLink_I interface {
	command(env string) *exec.Cmd
	run(cmd *exec.Cmd, gpuCount int, query string, action Action) ([]common.MapStr, error)
}

//BDLink implements one flavour of BDLink interface.
type BDLink struct {
}

//newBDLink returns instance of BDLink
func newBDLink() BDLink {
	return BDLink{}
}

func (g BDLink) command() *exec.Cmd {
	cmd := "ls -ln /opt/bluedata/dev | grep ^l | awk '{print $9\",\"$11}'"
	return exec.Command("bash", "-c", cmd)
}

//Run the ls command to collect bd GPU dev links
//Parse output and return array of link names.
func (g BDLink) run(cmd *exec.Cmd, action Action) (map[int]string, error) {
	reader := action.start(cmd)
	re := regexp.MustCompile(`\d+\z`)
	links := make(map[int]string)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		
		// No links -> GPUs not assigned
		if len(line) == 0 {
			logp.Debug("nvidiagpubeat: %s", "No links found")
			return nil, nil
		}

		logp.Debug("nvidiagpubeat", "Got Line %s",line)
		devLink := strings.Split(line, ",")
		devLinkName := strings.TrimSpace(devLink[0])
		gpuNumStr := re.FindString(strings.TrimSpace(devLink[1]))
		if len(gpuNumStr) == 0 {
			logp.Warn("nvidiagpubeat", "Bad formatted device name: %s", devLink[1])
			continue
		}
		gpuIndex, _ := strconv.Atoi(gpuNumStr)
		logp.Debug("nvidiagpubeat", "Got DevLinkName: %s, GPUDev %s, GPUIndex %d", devLinkName, devLink[1], gpuIndex)
		links[gpuIndex] = devLinkName

	}
	cmd.Wait()
	return links, nil
}
