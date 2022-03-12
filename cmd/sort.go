/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var infilename string
var outfilename string

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sort YAML keys",
	Run: func(cmd *cobra.Command, args []string) {
		var yamlFile []byte
		yamlFile, err := ioutil.ReadFile(infilename)
		if err != nil {
			log.Fatalln(err)
		}
		var yamlContents map[string]interface{}
		err = yaml.Unmarshal(yamlFile, &yamlContents)
		if err != nil {
			log.Fatalln(err)
		}
		// fmt.Printf("%#v", yamlContents)
		outBuffer, err := yaml.Marshal(yamlContents)
		if err != nil {
			log.Fatalln(err)
		}

		if err := ioutil.WriteFile(outfilename, outBuffer, 0644); err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sortCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	sortCmd.Flags().StringVar(&infilename, "in", "", "Input filename")
	sortCmd.MarkFlagRequired("in")
	sortCmd.Flags().StringVar(&outfilename, "out", "", "Output filename")
}
