// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	yaml "github.com/mobiledgex/yaml/v2"
)

var replace = flag.Bool("replace", false, "replace original file")

func main() {
	flag.Parse()
	// all args should be file names
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Please specify appdata yaml files to sort")
	}

	for _, inFile := range args {
		err := sortAppData(inFile)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func sortAppData(inFile string) error {
	dat, err := ioutil.ReadFile(inFile)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %s", inFile, err.Error())
	}
	appData := ormapi.AllData{}
	err = yaml.UnmarshalStrict(dat, &appData)
	if err != nil {
		return fmt.Errorf("Error parsing yaml for file %s: %s", inFile, err.Error())
	}
	appData.Sort()
	odat, err := yaml.Marshal(appData)
	if err != nil {
		return fmt.Errorf("Error marshaling yaml for file %s: %s", inFile, err.Error())
	}
	outFile := inFile
	if !*replace {
		outFile += ".sorted"
	}
	err = ioutil.WriteFile(outFile, odat, 0644)
	if err != nil {
		return fmt.Errorf("Error writing sorted data for file %s: %s", inFile, err.Error())
	}
	return nil
}
